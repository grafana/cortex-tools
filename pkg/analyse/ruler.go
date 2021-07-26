package analyse

type MetricsInRuler struct {
	MetricsUsed []string           `json:"metricsUsed"`
	RuleGroups  []RuleGroupMetrics `json:"ruleGroups"`
}

type RuleGroupMetrics struct {
	Namespace   string   `json:"namspace"`
	GroupName   string   `json:"name"`
	Metrics     []string `json:"metrics"`
	ParseErrors []string `json:"parse_errors"`
}
