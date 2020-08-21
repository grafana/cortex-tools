module github.com/grafana/cortex-tools

go 1.13

require (
	cloud.google.com/go/bigtable v1.2.0
	cloud.google.com/go/storage v1.6.0
	github.com/alecthomas/chroma v0.7.0
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/cortexproject/cortex v1.2.0
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/snappy v0.0.1
	github.com/google/go-cmp v0.4.0
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/opentracing/opentracing-go v1.1.1-0.20200124165624-2876d2018785
	github.com/pkg/errors v0.9.1
	github.com/prometheus/alertmanager v0.20.0
	github.com/prometheus/client_golang v1.6.1-0.20200604110148-03575cad4e55
	github.com/prometheus/common v0.10.0
	github.com/prometheus/prometheus v1.8.2-0.20200622142935-153f859b7499
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/testify v1.5.1
	google.golang.org/api v0.26.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200603094226-e3079894b1e8
)

// Cortex Overrides
replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v36.2.0+incompatible

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.0+incompatible

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0

replace k8s.io/client-go => k8s.io/client-go v0.18.3

// Use fork of gocql that has gokit logs and Prometheus metrics.
replace github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85
