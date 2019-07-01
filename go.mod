module github.com/grafana/ruler_cli

go 1.12

require (
	github.com/cortexproject/cortex v1.17.0
	github.com/prometheus/prometheus v0.0.0-20190417125241-3cc5f9d88062
	github.com/sirupsen/logrus v1.2.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/cortexproject/cortex => github.com/grafana/cortex v0.0.0-20190627165620-c56fa2946dac
