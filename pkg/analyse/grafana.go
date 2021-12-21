package analyse

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/grafana-tools/sdk"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/promql/parser"
)

type MetricsInGrafana struct {
	MetricsUsed    []string            `json:"metricsUsed"`
	OverallMetrics map[string]struct{} `json:"-"`
	Dashboards     []DashboardMetrics  `json:"dashboards"`
}

type DashboardMetrics struct {
	Slug        string   `json:"slug"`
	UID         string   `json:"uid,omitempty"`
	Title       string   `json:"title"`
	Metrics     []string `json:"metrics"`
	ParseErrors []string `json:"parse_errors"`
}

func ParseMetricsInBoard(mig *MetricsInGrafana, board sdk.Board) {
	var parseErrors []error
	metrics := make(map[string]struct{})

	// Iterate through all the panels and collect metrics
	for _, panel := range board.Panels {
		parseErrors = append(parseErrors, metricsFromPanel(*panel, metrics)...)
		if panel.RowPanel != nil {
			for _, subPanel := range panel.RowPanel.Panels {
				parseErrors = append(parseErrors, metricsFromPanel(subPanel, metrics)...)
			}
		}
	}

	// Iterate through all the rows and collect metrics
	for _, row := range board.Rows {
		for _, panel := range row.Panels {
			parseErrors = append(parseErrors, metricsFromPanel(panel, metrics)...)
		}
	}

	// Process metrics in templating
	parseErrors = append(parseErrors, metricsFromTemplating(board.Templating, metrics)...)

	var parseErrs []string
	for _, err := range parseErrors {
		parseErrs = append(parseErrs, err.Error())
	}

	var metricsInBoard []string
	for metric := range metrics {
		if metric == "" {
			continue
		}

		metricsInBoard = append(metricsInBoard, metric)
		mig.OverallMetrics[metric] = struct{}{}
	}
	sort.Strings(metricsInBoard)

	mig.Dashboards = append(mig.Dashboards, DashboardMetrics{
		Slug:        board.Slug,
		UID:         board.UID,
		Title:       board.Title,
		Metrics:     metricsInBoard,
		ParseErrors: parseErrs,
	})

}

func metricsFromTemplating(templating sdk.Templating, metrics map[string]struct{}) []error {
	parseErrors := []error{}
	for _, templateVar := range templating.List {
		if templateVar.Type != "query" {
			continue
		}

		query, ok := templateVar.Query.(string)
		if !ok {
			iter := reflect.ValueOf(templateVar.Query).MapRange() // A query struct
			for iter.Next() {
				key := iter.Key().Interface()
				value := iter.Value().Interface()
				if key == "query" {
					query = value.(string)
					break
				}
			}
		}

		if query != "" {
			// label_values
			if strings.Contains(query, "label_values") {
				re := regexp.MustCompile(`label_values\(\s*([a-zA-Z0-9_]+)`)
				sm := re.FindStringSubmatch(query)
				// In case of really gross queries, like - https://github.com/grafana/jsonnet-libs/blob/e97ab17f67ab40d5fe3af7e59151dd43be03f631/hass-mixin/dashboard.libsonnet#L93
				if len(sm) > 0 {
					query = sm[1]
				}
			}
			// query_result
			if strings.Contains(query, "query_result") {
				re := regexp.MustCompile(`query_result\((.+)\)`)
				query = re.FindStringSubmatch(query)[1]
			}
			err := parseQuery(query, metrics)
			if err != nil {
				parseErrors = append(parseErrors, errors.Wrapf(err, "query=%v", query))
				log.Debugln("msg", "promql parse error", "err", err, "query", query)
				continue
			}
		} else {
			err := fmt.Errorf("templating type error: name=%v", templateVar.Name)
			parseErrors = append(parseErrors, err)
			log.Debugln("msg", "templating parse error", "err", err)
			continue
		}
	}
	return parseErrors
}

// Workaround to support Grafana "timeseries" panel. This should
// be implemented in grafana/tools-sdk, and removed from here.
func getCustomPanelTargets(panel sdk.Panel) *[]sdk.Target {

	switch panel.CommonPanel.Type {
	case
		"timeseries",
		"piechart",
		"gauge",
		"table-old",
		"jdbranham-diagram-panel",
		"agenty-flowcharting-panel",
		"grafana-worldmap-panel",
		"digiapulssi-breadcrumb-panel",
		"grafana-piechart-panel":
	default:
		return nil
	}

	// Heavy handed approach to re-marshal the panel and parse it again
	// so that we can extract the 'targets' field in the right format.

	bytes, err := json.Marshal(panel.CustomPanel)
	if err != nil {
		log.Debugln("msg", "panel re-marshalling error", "err", err)
		return nil
	}

	type panelType struct {
		Targets []sdk.Target `json:"targets,omitempty"`
	}

	var parsedPanel panelType
	err = json.Unmarshal(bytes, &parsedPanel)
	if err != nil {
		log.Debugln("msg", "panel parsing error", "err", err)
		return nil
	}

	return &parsedPanel.Targets
}

func metricsFromPanel(panel sdk.Panel, metrics map[string]struct{}) []error {
	var parseErrors []error

	switch panel.CommonPanel.Type {
	case
		"row",
		"welcome",
		"dashlist",
		"news",
		"annolist",
		"alertlist",
		"pluginlist",
		"grafana-clock-panel",
		"text":
		return parseErrors // Let's not throw parse errors...these don't contain queries!
	}

	targets := panel.GetTargets()
	if targets == nil {
		targets = getCustomPanelTargets(panel)
		if targets == nil {
			parseErrors = append(parseErrors, fmt.Errorf("unsupported panel type: %q", panel.CommonPanel.Type))
			return parseErrors
		}
	}

	for _, target := range *targets {
		// Prometheus has this set.
		if target.Expr == "" {
			continue
		}
		query := target.Expr
		err := parseQuery(query, metrics)
		if err != nil {
			parseErrors = append(parseErrors, errors.Wrapf(err, "query=%v", query))
			log.Debugln("msg", "promql parse error", "err", err, "query", query)
			continue
		}
	}

	return parseErrors
}

func parseQuery(query string, metrics map[string]struct{}) error {
	re := regexp.MustCompile(`\[\s*\$(\w+|{\w+})\]`) // variable rate interval
	query = re.ReplaceAllString(query, "[1s]")

	re1 := regexp.MustCompile(`offset\s+\$(\w+|{\w+})`) // variable offset
	query = re1.ReplaceAllString(query, "offset 1s")

	re2 := regexp.MustCompile(`(by\s*\()\$((\w+|{\w+}))`) // variable by clause
	query = re2.ReplaceAllString(query, "$1$2")

	query = strings.ReplaceAll(query, `$__interval`, "5m")
	query = strings.ReplaceAll(query, `$interval`, "5m")
	query = strings.ReplaceAll(query, `$resolution`, "5s")
	query = strings.ReplaceAll(query, "$__rate_interval", "15s")
	query = strings.ReplaceAll(query, "$__range", "1d")
	query = strings.ReplaceAll(query, "${__range_s:glob}", "30")
	query = strings.ReplaceAll(query, "${__range_s}", "30")

	re3 := regexp.MustCompile(`\$(__(to|from):date:\w+\b|{__(to|from):date:\w+})`)
	query = re3.ReplaceAllString(query, "12") // Replace dates

	// Replace *all* variables.  Magic number 79197919 is used as it's unlikely to appear in a query.  Some queries have
	// variable metric names. There is a check a few lines below for this edge case.
	re4 := regexp.MustCompile(`\$(\w+|{\w+})`)
	query = re4.ReplaceAllString(query, "79197919")

	expr, err := parser.ParseExpr(query)
	if err != nil {
		return err
	}

	parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
		if n, ok := node.(*parser.VectorSelector); ok {
			if strings.Contains(n.Name, "79197919") { // Check for the magic number in the metric name..drop it
				return errors.New("Query contains a variable in the metric name")
			}
			metrics[n.Name] = struct{}{}
		}

		return nil
	})

	return nil
}
