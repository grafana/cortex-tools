module github.com/grafana/cortex-tools

go 1.14

require (
	cloud.google.com/go/bigtable v1.2.0
	cloud.google.com/go/storage v1.10.0
	github.com/alecthomas/chroma v0.7.0
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15
	github.com/cortexproject/cortex v1.7.1-0.20210310163109-5d23b16190f6
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/go-kit/kit v0.10.0
	github.com/gocql/gocql v0.0.0-20200526081602-cd04bd7f22a7
	github.com/gogo/protobuf v1.3.2
	github.com/golang/snappy v0.0.3-0.20201103224600-674baa8c7fc3
	github.com/gonum/blas v0.0.0-20181208220705-f22b278b28ac // indirect
	github.com/gonum/floats v0.0.0-20181209220543-c233463c7e82 // indirect
	github.com/gonum/integrate v0.0.0-20181209220457-a422b5c0fdf2 // indirect
	github.com/gonum/internal v0.0.0-20181124074243-f884aa714029 // indirect
	github.com/gonum/lapack v0.0.0-20181123203213-e4cdc5a0bff9 // indirect
	github.com/gonum/matrix v0.0.0-20181209220409-c518dec07be9 // indirect
	github.com/gonum/stat v0.0.0-20181125101827-41a0da705a5b
	github.com/google/go-github/v32 v32.1.0
	github.com/gorilla/mux v1.7.3
	github.com/grafana-tools/sdk v0.0.0-20210310213032-c3f3511b3e9b
	github.com/grafana/loki v1.6.2-0.20210310125813-306cc724380c
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/alertmanager v0.21.1-0.20210310093010-0f9cab6991e6
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/common v0.18.0
	github.com/prometheus/prometheus v1.8.2-0.20210215121130-6f488061dfb4
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/weaveworks/common v0.0.0-20210112142934-23c8d7fa6120
	go.uber.org/atomic v1.7.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	google.golang.org/api v0.39.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools v2.2.0+incompatible
)

// Cortex Overrides
replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v36.2.0+incompatible

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/satori/go.uuid => github.com/satori/go.uuid v1.2.0

// Keeping this same as Cortex to avoid dependency issues.
replace k8s.io/client-go => k8s.io/client-go v0.20.4

replace k8s.io/api => k8s.io/api v0.20.4

// Use fork of gocql that has gokit logs and Prometheus metrics.
replace github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85

// Using a 3rd-party branch for custom dialer - see https://github.com/bradfitz/gomemcache/pull/86
replace github.com/bradfitz/gomemcache => github.com/themihai/gomemcache v0.0.0-20180902122335-24332e2d58ab

// Required for Alertmanager
replace github.com/go-openapi/errors => github.com/go-openapi/errors v0.19.4

replace github.com/go-openapi/validate => github.com/go-openapi/validate v0.19.8

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.8.1
