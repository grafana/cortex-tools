package bench

import (
	"context"
	"flag"
	"sync"

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
	cfg WriteBenchConfig

	// Do DNS client side load balancing if configured
	remoteMtx  sync.Mutex
	addresses  []string
	clientPool map[string]*writeClient

	workload *workload

	reg    prometheus.Registerer
	logger log.Logger

	requestDuration *prometheus.HistogramVec
}

func newQueryRunner(id string, cfg WriteBenchConfig, workload *workload, logger log.Logger, reg prometheus.Registerer) (*queryRunner, error) {
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
	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}
