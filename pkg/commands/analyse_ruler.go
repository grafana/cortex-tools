package commands

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"

	"github.com/grafana/cortex-tools/pkg/analyse"
	"github.com/grafana/cortex-tools/pkg/client"
	"github.com/grafana/cortex-tools/pkg/rules/rwrulefmt"
	"github.com/prometheus/prometheus/promql/parser"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type RulerAnalyseCommand struct {
	ClientConfig client.Config

	cli *client.CortexClient

	outputFile string
}

func (cmd *RulerAnalyseCommand) run(k *kingpin.ParseContext) error {
	output := analyse.MetricsInRuler{}
	allMetrics := map[string]struct{}{}

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
			metrics, err := parseMetricsInRuleGroup(ns, rg)
			// todo
			if err != nil {
				log.Fatalf("metrics parse error %v", err)
			}

			metricsInGroup := make([]string, 0, len(metrics))
			for metric := range metrics {
				if metric == "" {
					continue
				}
				metricsInGroup = append(metricsInGroup, metric)
				allMetrics[metric] = struct{}{}
			}
			output.RuleGroups = append(output.RuleGroups, analyse.RuleGroupMetrics{
				Namespace: ns,
				GroupName: rg.Name,
				Metrics:   metricsInGroup,
			})
		}
	}

	metricsUsed := make([]string, 0, len(allMetrics))
	for metric := range allMetrics {
		metricsUsed = append(metricsUsed, metric)
	}
	sort.Strings(metricsUsed)

	output.MetricsUsed = metricsUsed
	out, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(cmd.outputFile, out, os.FileMode(int(0666))); err != nil {
		return err
	}

	return nil
}

func parseMetricsInRuleGroup(ns string, group rwrulefmt.RuleGroup) (map[string]struct{}, error) {
	ruleMetrics := map[string]struct{}{}
	refMetrics := map[string]struct{}{}

	rules := group.Rules
	for _, rule := range rules {
		if rule.Record.Value != "" {
			ruleMetrics[rule.Record.Value] = struct{}{}
		} else {
			ruleMetrics[rule.Alert.Value] = struct{}{}
		}

		query := rule.Expr.Value
		expr, err := parser.ParseExpr(query)
		// todo maintain parse errors
		if err != nil {
			return refMetrics, err
		}

		parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
			if n, ok := node.(*parser.VectorSelector); ok {
				refMetrics[n.Name] = struct{}{}
			}

			return nil
		})
	}

	// should we handle metrics referenced in other RGs?
	metrics := diff(ruleMetrics, refMetrics)
	return metrics, nil
}

func diff(ruleMetrics map[string]struct{}, refMetrics map[string]struct{}) map[string]struct{} {
	for ruleMetric := range ruleMetrics {
		if _, ok := refMetrics[ruleMetric]; ok {
			delete(refMetrics, ruleMetric)
		}
	}
	return refMetrics
}
