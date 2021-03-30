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

	"github.com/cortexproject/cortex/pkg/util/spanlogger"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	config_util "github.com/prometheus/common/config"
	"github.com/thanos-io/thanos/pkg/discovery/dns"
)

type QueryConfig struct {
	Enabled           bool   `yaml:"enabled"`
	Endpoint          string `yaml:"endpoint"`
	BasicAuthUsername string `yaml:"basic_auth_username"`
	BasicAuthPasword  string `yaml:"basic_auth_password"`
}

func (cfg *QueryConfig) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&cfg.Enabled, "bench.query.enabled", false, "enable query benchmarking")
	f.StringVar(&cfg.Endpoint, "bench.query.endpoint", "", "Remote query endpoint.")
	f.StringVar(&cfg.BasicAuthUsername, "bench.query.basic-auth-username", "", "Set the basic auth username on remote query requests.")
	f.StringVar(&cfg.BasicAuthPasword, "bench.query.basic-auth-password", "", "Set the basic auth password on remote query requests.")
}

type queryRunner struct {
	id  string
	cfg QueryConfig

	// Do DNS client side load balancing if configured
	dnsProvider *dns.Provider
	remoteMtx   sync.Mutex
	addresses   []string
	clientPool  map[string]v1.API

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
		clientPool: map[string]v1.API{},
		logger:     logger,
		reg:        reg,
		requestDuration: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "benchtool",
				Name:      "query_request_duration_seconds",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120},
			},
			[]string{"code", "type"},
		),
	}

	return runner, nil
}

func (q *queryRunner) Run(ctx context.Context) error {
	queryChan := make(chan query, 50)
	for i := 0; i < 50; i++ {
		go q.queryWorker(queryChan)
	}
	for _, queryReq := range q.workload.queries {
		// every query has a ticker and a Go loop...
		// not sure if this is a good idea but it should be fine
		go func(q query) {
			ticker := time.NewTicker(q.interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					queryChan <- q
				case <-ctx.Done():
					return
				}
			}
		}(queryReq)
	}
	for {
		select {
		case <-ctx.Done():
			close(queryChan)
			return nil
		}
	}
}

func (q *queryRunner) queryWorker(queryChan chan query) {
	for queryReq := range queryChan {
		err := q.executeQuery(context.Background(), queryReq)
		if err != nil {
			level.Warn(q.logger).Log("msg", "unable to execute query", "err", err)
		}
	}
}

func newQueryClient(url, username, password string) (v1.API, error) {
	apiClient, err := api.NewClient(api.Config{
		Address:      url,
		RoundTripper: config_util.NewBasicAuthRoundTripper(username, config_util.Secret(password), "", api.DefaultRoundTripper),
	})

	if err != nil {
		return nil, err
	}
	return v1.NewAPI(apiClient), nil
}

func (w *queryRunner) getRandomAPIClient() (v1.API, error) {
	w.remoteMtx.Lock()
	defer w.remoteMtx.Unlock()

	randomIndex := rand.Intn(len(w.addresses))
	pick := w.addresses[randomIndex]

	var cli v1.API
	var exists bool

	if cli, exists = w.clientPool[pick]; !exists {
		cli, err := newQueryClient("http://"+pick+"/prometheus", w.cfg.BasicAuthUsername, w.cfg.BasicAuthPasword)
		if err != nil {
			return nil, err
		}
		w.clientPool[pick] = cli
	}

	return cli, nil
}

func (q *queryRunner) executeQuery(ctx context.Context, queryReq query) error {
	spanLog, ctx := spanlogger.New(ctx, "queryRunner.executeQuery")
	defer spanLog.Span.Finish()
	apiClient, err := q.getRandomAPIClient()
	if err != nil {
		return err
	}

	// Create a timestamp for use when creating the requests and observing latency
	now := time.Now()

	var (
		queryType string = "instant"
		status    string = "success"
	)
	if queryReq.timeRange > 0 {
		queryType = "range"
		level.Debug(q.logger).Log("msg", "sending range query", "expr", queryReq.expr, "range", queryReq.timeRange)
		r := v1.Range{
			Start: now.Add(-queryReq.timeRange),
			End:   now,
			Step:  time.Minute,
		}
		_, _, err = apiClient.QueryRange(ctx, queryReq.expr, r)
	} else {
		level.Debug(q.logger).Log("msg", "sending instant query", "expr", queryReq.expr)
		_, _, err = apiClient.Query(ctx, queryReq.expr, now)
	}
	if err != nil {
		status = "failure"
	}

	q.requestDuration.WithLabelValues(status, queryType).Observe(time.Since(now).Seconds())
	return err
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
		GaugeZero:     nil,
		GaugeRandom:   nil,
		CounterOne:    nil,
		CounterRandom: nil,
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

func (q *queryRunner) resolveAddrsLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := q.resolveAddrs()
			if err != nil {
				level.Warn(q.logger).Log("msg", "failed update remote write servers list", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (q *queryRunner) resolveAddrs() error {
	// Resolve configured addresses with a reasonable timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// If some of the dns resolution fails, log the error.
	if err := q.dnsProvider.Resolve(ctx, []string{q.cfg.Endpoint}); err != nil {
		level.Error(q.logger).Log("msg", "failed to resolve addresses", "err", err)
	}

	// Fail in case no server address is resolved.
	servers := q.dnsProvider.Addresses()
	if len(servers) == 0 {
		return errors.New("no server address resolved")
	}

	q.remoteMtx.Lock()
	q.addresses = servers
	q.remoteMtx.Unlock()

	return nil
}
