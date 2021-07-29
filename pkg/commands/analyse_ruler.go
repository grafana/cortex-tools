package commands

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"

	"github.com/prometheus/prometheus/promql/parser"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/grafana/cortex-tools/pkg/analyse"
	"github.com/grafana/cortex-tools/pkg/client"
	"github.com/grafana/cortex-tools/pkg/rules/rwrulefmt"
)

type RulerAnalyseCommand struct {
	ClientConfig client.Config
	cli          *client.CortexClient
	outputFile   string
}

func (cmd *RulerAnalyseCommand) run(k *kingpin.ParseContext) error {
	output := &analyse.MetricsInRuler{}
	output.OverallMetrics = make(map[string]struct{})

	cli, err := client.New(cmd.ClientConfig)
	if err != nil {
		return err
	}

	cmd.cli = cli
	rules, err := cmd.cli.ListRules(context.Background(), "")
	if err != nil {
		log.Fatalf("unable to read rules from cortex, %v", err)
	}

	for ns := range rules {
		for _, rg := range rules[ns] {
			err := parseMetricsInRuleGroup(output, rg, ns)
			// todo
			if err != nil {
				log.Fatalf("metrics parse error %v", err)
			}
		}
	}

	err = writeOutRuleMetrics(output, cmd.outputFile)
	if err != nil {
		return err
	}

	return nil
}

func parseMetricsInRuleGroup(mir *analyse.MetricsInRuler, group rwrulefmt.RuleGroup, ns string) error {
	ruleMetrics := map[string]struct{}{}
	refMetrics := map[string]struct{}{}

	rules := group.Rules
	for _, rule := range rules {
		//todo check this
		if rule.Record.Value != "" {
			ruleMetrics[rule.Record.Value] = struct{}{}
		}

		query := rule.Expr.Value
		expr, err := parser.ParseExpr(query)
		// todo maintain parse errors
		if err != nil {
			return err
		}

		parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
			if n, ok := node.(*parser.VectorSelector); ok {
				refMetrics[n.Name] = struct{}{}
			}

			return nil
		})
	}

	// remove defined recording rule metrics in same RG
	metrics := diff(ruleMetrics, refMetrics)

	metricsInGroup := make([]string, 0, len(metrics))
	for metric := range metrics {
		if metric == "" {
			continue
		}
		metricsInGroup = append(metricsInGroup, metric)
		mir.OverallMetrics[metric] = struct{}{}
	}
	mir.RuleGroups = append(mir.RuleGroups, analyse.RuleGroupMetrics{
		Namespace: ns,
		GroupName: group.Name,
		Metrics:   metricsInGroup,
	})

	return nil
}

func diff(ruleMetrics map[string]struct{}, refMetrics map[string]struct{}) map[string]struct{} {
	for ruleMetric := range ruleMetrics {
		if _, ok := refMetrics[ruleMetric]; ok {
			delete(refMetrics, ruleMetric)
		}
	}
	return refMetrics
}

func writeOutRuleMetrics(mir *analyse.MetricsInRuler, outputFile string) error {
	metricsUsed := make([]string, 0, len(mir.OverallMetrics))
	for metric := range mir.OverallMetrics {
		metricsUsed = append(metricsUsed, metric)
	}
	sort.Strings(metricsUsed)

	mir.MetricsUsed = metricsUsed
	out, err := json.MarshalIndent(mir, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputFile, out, os.FileMode(int(0666))); err != nil {
		return err
	}

	return nil
}
