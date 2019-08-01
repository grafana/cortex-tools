module github.com/grafana/cortex-tool

go 1.12

// TODO: Temporary until rulesdb refactor is merged https://github.com/cortexproject/cortex/pull/1513
replace github.com/cortexproject/cortex => ../cortex

require (
	cloud.google.com/go v0.35.0
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/cortexproject/cortex v0.0.0-00010101000000-000000000000
	github.com/go-kit/kit v0.8.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.4.1
	github.com/prometheus/prometheus v0.0.0-20190417125241-3cc5f9d88062
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	github.com/weaveworks/common v0.0.0-20190410110702-87611edc252e
	golang.org/x/arch v0.0.0-20190312162104-788fe5ffcd8c // indirect
	google.golang.org/api v0.4.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.2
)
