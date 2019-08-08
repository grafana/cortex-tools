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
	github.com/prometheus/prometheus v0.0.0-20190731144842-63ed2e28f1ac
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	github.com/weaveworks/common v0.0.0-20190714171817-ddeaa31513fd
	golang.org/x/arch v0.0.0-20190312162104-788fe5ffcd8c // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80 // indirect
	golang.org/x/sys v0.0.0-20190804053845-51ab0e2deafa // indirect
	golang.org/x/tools v0.0.0-20190807164442-cae9aa543496 // indirect
	golang.org/x/tools/gopls v0.1.4-0.20190807164442-cae9aa543496 // indirect
	google.golang.org/api v0.4.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.2
)
