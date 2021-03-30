package bench

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"hash/adler32"
	"html/template"
	"math/rand"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type QueryConfig struct {
	Enabled           bool   `yaml:"enabled"`
	Endpoint          string `yaml:"endpoint"`
	BasicAuthUsername string `yaml:"basic_auth_username"`
	BasicAuthPasword  string `yaml:"basic_auth_password"`
}

func (cfg *QueryConfig) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&cfg.Enabled, "bench.query.enabled", true, "enable query benchmarking")
	f.StringVar(&cfg.Endpoint, "bench.query.endpoint", "", "Remote query endpoint.")
	f.StringVar(&cfg.BasicAuthUsername, "bench.query.basic-auth-username", "", "Set the basic auth username on remote query requests.")
	f.StringVar(&cfg.BasicAuthPasword, "bench.query.basic-auth-password", "", "Set the basic auth password on remote query requests.")
}

type queryRunner struct {
	id  string
	cfg QueryConfig

	// Do DNS client side load balancing if configured
	remoteMtx  sync.Mutex
	addresses  []string
	clientPool map[string]*writeClient

	workload *queryWorkload

	reg    prometheus.Registerer
	logger log.Logger

	requestDuration *prometheus.HistogramVec
}

func newQueryRunner(id string, cfg QueryConfig, workload *queryWorkload, logger log.Logger, reg prometheus.Registerer) (*queryRunner, error) {
	runner := &queryRunner{
		id:  id,
		cfg: cfg,

		workload:   workload,
		clientPool: map[string]*writeClient{},
		logger:     logger,
		reg:        reg,
		requestDuration: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "benchtool",
				Name:      "query_request_duration_seconds",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120},
			},
			[]string{"code"},
		),
	}

	return runner, nil
}

func (q *queryRunner) Run(ctx context.Context) error {
	//	for _, query := range q.workload.queries {
	// TODO: start a go function for each query to be sent to a channel
	//}
	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}

type query struct {
	interval  time.Duration
	timeRange time.Duration
	expr      string
}

type queryWorkload struct {
	queries []query
}

type exprTemplateData struct {
	Name     string
	Matchers string
}

func newQueryWorkload(id string, desc WorkloadDesc) (*queryWorkload, error) {
	seriesTypeMap := map[SeriesType][]SeriesDesc{
		GaugeZero:     []SeriesDesc{},
		GaugeRandom:   []SeriesDesc{},
		CounterOne:    []SeriesDesc{},
		CounterRandom: []SeriesDesc{},
	}

	for _, s := range desc.Series {
		seriesSlice, ok := seriesTypeMap[s.Type]
		if !ok {
			return nil, fmt.Errorf("series found with unknown series type %s", s.Type)
		}

		seriesTypeMap[s.Type] = append(seriesSlice, s)
	}

	// Use the provided ID to create a random seed. This will ensure repeated runs with the same
	// configured ID will produce the same query workloads.
	hashSeed := adler32.Checksum([]byte(id))
	rand := rand.New(rand.NewSource(int64(hashSeed)))

	queries := []query{}
	for _, queryDesc := range desc.QueryDesc {
		exprTemplate, err := template.New("query").Parse(queryDesc.ExprTemplate)
		if err != nil {
			return nil, fmt.Errorf("unable to parse query template, %v", err)
		}

		for i := 0; i < queryDesc.NumQueries; i++ {
			seriesSlice, ok := seriesTypeMap[queryDesc.RequiredSeriesType]
			if !ok {
				return nil, fmt.Errorf("query found with unknown series type %s", queryDesc.RequiredSeriesType)
			}

			if len(seriesSlice) == 0 {
				return nil, fmt.Errorf("no series found for query with series type %s", queryDesc.RequiredSeriesType)
			}

			seriesDesc := seriesSlice[rand.Intn(len(seriesSlice))]

			var b bytes.Buffer
			exprTemplate.Execute(&b, exprTemplateData{
				Name: seriesDesc.Name,
			})

			queries = append(queries, query{
				interval:  queryDesc.Interval,
				timeRange: queryDesc.TimeRange,
				expr:      b.String(),
			})
		}
	}

	return &queryWorkload{queries}, nil
}
