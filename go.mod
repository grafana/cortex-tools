module github.com/grafana/cortex-tools

go 1.16

require (
	cloud.google.com/go/bigtable v1.3.0
	cloud.google.com/go/storage v1.10.0
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/alecthomas/chroma v0.7.0
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/alecthomas/units v0.0.0-20210912230133-d1bdfacee922
	github.com/aws/aws-lambda-go v1.17.0 // indirect
	github.com/cortexproject/cortex v1.10.1-0.20211104100946-3f329a21cad4
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/drone/envsubst v1.0.2 // indirect
	github.com/go-kit/kit v0.12.0
	github.com/gocql/gocql v0.0.0-20200526081602-cd04bd7f22a7
	github.com/gogo/protobuf v1.3.2
	github.com/golang/snappy v0.0.4
	github.com/gonum/blas v0.0.0-20181208220705-f22b278b28ac // indirect
	github.com/gonum/floats v0.0.0-20181209220543-c233463c7e82 // indirect
	github.com/gonum/integrate v0.0.0-20181209220457-a422b5c0fdf2 // indirect
	github.com/gonum/internal v0.0.0-20181124074243-f884aa714029 // indirect
	github.com/gonum/lapack v0.0.0-20181123203213-e4cdc5a0bff9 // indirect
	github.com/gonum/matrix v0.0.0-20181209220409-c518dec07be9 // indirect
	github.com/gonum/stat v0.0.0-20181125101827-41a0da705a5b
	github.com/google/go-github/v32 v32.1.0
	github.com/gorilla/mux v1.8.0
	github.com/grafana-tools/sdk v0.0.0-20210808170449-9de4d14888c5
	github.com/grafana/dskit v0.0.0-20211103155626-4e784973d341
	github.com/grafana/loki v1.6.2-0.20211108134117-5b9f5b9efaa5
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.1-0.20191002090509-6af20e3a5340 // indirect
	github.com/influxdata/go-syslog/v3 v3.0.1-0.20201128200927-a1889d947b48 // indirect
	github.com/influxdata/telegraf v1.16.3 // indirect
	github.com/klauspost/pgzip v1.2.5 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/oklog/ulid v1.3.1
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/alertmanager v0.23.1-0.20210914172521-e35efbddb66a
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.31.1
	github.com/prometheus/prometheus v1.8.2-0.20211011171444-354d8d2ecfac
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/thanos-io/thanos v0.22.0
	github.com/weaveworks/common v0.0.0-20211015155308-ebe5bdc2c89e
	go.opentelemetry.io/otel v0.20.0 // indirect
	go.uber.org/atomic v1.9.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/api v0.57.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools v2.2.0+incompatible
)

// Cortex Overrides
replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
replace github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.40.37
replace github.com/thanos-io/thanos v0.22.0 => github.com/thanos-io/thanos v0.19.1-0.20210923155558-c15594a03c45

// Thanos Override
replace github.com/efficientgo/tools/core => github.com/efficientgo/tools/core v0.0.0-20210731122119-5d4a0645ce9a

// Keeping this same as Cortex to avoid dependency issues.
replace k8s.io/client-go => k8s.io/client-go v0.20.4

replace k8s.io/api => k8s.io/api v0.20.4

// Use fork of gocql that has gokit logs and Prometheus metrics.
replace github.com/gocql/gocql => github.com/grafana/gocql v0.0.0-20200605141915-ba5dc39ece85

// Using a 3rd-party branch for custom dialer - see https://github.com/bradfitz/gomemcache/pull/86
replace github.com/bradfitz/gomemcache => github.com/themihai/gomemcache v0.0.0-20180902122335-24332e2d58ab

// Required for Alertmanager

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.8.1
