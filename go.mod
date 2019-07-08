module github.com/grafana/cortex-tool

go 1.12

// Temporary until rulesdb refactor is merged
replace github.com/cortexproject/cortex => ../cortex

require (
	github.com/cortexproject/cortex v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.8.1
	github.com/prometheus/prometheus v0.0.0-20190417125241-3cc5f9d88062
	github.com/sirupsen/logrus v1.4.2
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.2
)
