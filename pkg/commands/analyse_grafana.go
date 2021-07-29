package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/grafana-tools/sdk"
	"github.com/grafana/cortex-tools/pkg/analyse"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/promql/parser"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type GrafanaAnalyseCommand struct {
	address     string
	apiKey      string
	readTimeout time.Duration

	outputFile string
}

func (cmd *GrafanaAnalyseCommand) run(k *kingpin.ParseContext) error {
	ctx, cancel := context.WithTimeout(context.Background(), cmd.readTimeout)
	defer cancel()

	c := sdk.NewClient(cmd.address, cmd.apiKey, sdk.DefaultHTTPClient)
	boardLinks, err := c.SearchDashboards(ctx, "", false)
	if err != nil {
		return err
	}

	output := &analyse.MetricsInGrafana{}
	output.OverallMetrics = make(map[string]struct{})

	for _, link := range boardLinks {
		board, _, err := c.GetDashboardByUID(ctx, link.UID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s for %s %s\n", err, link.UID, link.Title)
			continue
		}
		parseMetricsInBoard(output, board)
	}

	err = writeOut(output, cmd.outputFile)
	if err != nil {
		return err
	}

	return nil
}

func writeOut(mig *analyse.MetricsInGrafana, outputFile string) error {
	metricsUsed := make([]string, 0, len(mig.OverallMetrics))
	for metric := range mig.OverallMetrics {
		metricsUsed = append(metricsUsed, metric)
	}
	sort.Strings(metricsUsed)

	mig.MetricsUsed = metricsUsed
	out, err := json.MarshalIndent(mig, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputFile, out, os.FileMode(int(0666))); err != nil {
		return err
	}

	return nil
}

func parseMetricsInBoard(mig *analyse.MetricsInGrafana, board sdk.Board) {
	metrics := map[string]struct{}{}
	parseErrors := make([]error, 0)

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

	parseErrs := make([]string, 0, len(parseErrors))
	for _, err := range parseErrors {
		parseErrs = append(parseErrs, err.Error())
	}

	metricsInBoard := make([]string, 0, len(metrics))
	for metric := range metrics {
		if metric == "" {
			continue
		}

		metricsInBoard = append(metricsInBoard, metric)
		mig.OverallMetrics[metric] = struct{}{}
	}
	sort.Strings(metricsInBoard)

	mig.Dashboards = append(mig.Dashboards, analyse.DashboardMetrics{
		Slug:        board.Slug,
		UID:         board.UID,
		Title:       board.Title,
		Metrics:     metricsInBoard,
		ParseErrors: parseErrs,
	})

}

func metricsFromPanel(panel sdk.Panel, metrics map[string]struct{}) []error {
	parseErrors := []error{}
	if panel.GetTargets() == nil {
		return parseErrors
	}

	for _, target := range *panel.GetTargets() {
		// Prometheus has this set.
		if target.Expr == "" {
			continue
		}

		query := target.Expr
		query = strings.ReplaceAll(query, `$__interval`, "5m")
		query = strings.ReplaceAll(query, `$interval`, "5m")
		query = strings.ReplaceAll(query, `$resolution`, "5s")
		query = strings.ReplaceAll(query, "$__rate_interval", "15s")
		query = strings.ReplaceAll(query, "$__range", "1d")
		query = strings.ReplaceAll(query, "${__range_s:glob}", "30")
		query = strings.ReplaceAll(query, "${__range_s}", "30")
		expr, err := parser.ParseExpr(query)
		if err != nil {
			parseErrors = append(parseErrors, errors.Wrapf(err, "query=%v", query))
			log.Debugln("msg", "promql parse error", "err", err, "query", query)
			continue
		}

		parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
			if n, ok := node.(*parser.VectorSelector); ok {
				metrics[n.Name] = struct{}{}
			}

			return nil
		})
	}

	return parseErrors
}
