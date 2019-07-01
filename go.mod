module github.com/grafana/cortex-tools

go 1.12

replace github.com/cortexproject/cortex => github.com/grafana/cortex v0.0.0-20190627165620-c56fa2946dac

require (
	github.com/cortexproject/cortex v0.0.0-00010101000000-000000000000
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/sirupsen/logrus v1.4.2
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.2
)
