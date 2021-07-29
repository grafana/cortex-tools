package analyse

import "github.com/grafana/cortex-tools/pkg/rules/rwrulefmt"

type MetricsInRuler struct {
	MetricsUsed    []string            `json:"metricsUsed"`
	OverallMetrics map[string]struct{} `json:"overallMetrics"`
	RuleGroups     []RuleGroupMetrics  `json:"ruleGroups"`
}

type RuleGroupMetrics struct {
	Namespace   string   `json:"namspace"`
	GroupName   string   `json:"name"`
	Metrics     []string `json:"metrics"`
	ParseErrors []string `json:"parse_errors"`
}

type RuleConfig struct {
	Namespace  string                `yaml:"namespace"`
	RuleGroups []rwrulefmt.RuleGroup `yaml:"groups"`
}
