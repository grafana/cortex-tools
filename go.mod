module github.com/grafana/cortextool

go 1.12

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1

require (
	cloud.google.com/go v0.44.1
	github.com/alecthomas/chroma v0.7.0
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/cortexproject/cortex v0.4.0
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/gogo/protobuf v1.2.2-0.20190730201129-28a6bbf47e48
	github.com/golang/snappy v0.0.1
	github.com/google/martian v2.1.0+incompatible
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/alertmanager v0.20.0
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/common v0.7.0
	github.com/prometheus/prometheus v1.8.2-0.20190918104050-8744afdd1ea0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	google.golang.org/api v0.8.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.2
	gopkg.in/yaml.v3 v3.0.0-20200506231410-2ff61e1afc86
	sigs.k8s.io/yaml v1.1.0
)
