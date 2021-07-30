package commands

import (
	"testing"

	"github.com/grafana/cortex-tools/pkg/analyse"
	"github.com/grafana/cortex-tools/pkg/rules"
	"github.com/stretchr/testify/assert"
)

func TestParseMetricsInRuleFile(t *testing.T) {
	var metrics = []string{
		"apiserver_request_duration_seconds_bucket",
		"apiserver_request_duration_seconds_count",
		"apiserver_request_total",
	}

	output := &analyse.MetricsInRuler{}
	output.OverallMetrics = make(map[string]struct{})

	nss, err := rules.ParseFiles("cortex", []string{"testdata/prometheus_rules.yaml"})
	if err != nil {
		t.Errorf("could not parse rules file")
	}

	for _, ns := range nss {
		for _, group := range ns.Groups {
			err := parseMetricsInRuleGroup(output, group, ns.Namespace)
			if err != nil {
				t.Errorf("could not parse metrics in rule group")
			}
		}
	}
	assert.Equal(t, metrics, output.RuleGroups[0].Metrics)
}
