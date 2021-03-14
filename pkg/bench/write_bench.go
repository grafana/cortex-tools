package bench

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	ingester_client "github.com/cortexproject/cortex/pkg/ingester/client"
	"github.com/cortexproject/cortex/pkg/ring"
	"github.com/cortexproject/cortex/pkg/ring/kv/codec"
	"github.com/cortexproject/cortex/pkg/ring/kv/memberlist"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/thanos-io/thanos/pkg/discovery/dns"
	"github.com/thanos-io/thanos/pkg/extprom"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ID               string `yaml:"id"`
	InstanceName     string `yaml:"instance_name"`
	WorkloadFilePath string `yaml:"workload_file_path"`

	RingCheck RingCheckConfig  `yaml:"ring_check"`
	Write     WriteBenchConfig `yaml:"writes"`
}

func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	defaultID, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	f.StringVar(&cfg.ID, "bench.id", defaultID, "ID of worker. Defaults to hostname")
	f.StringVar(&cfg.InstanceName, "bench.instance-name", "default", "Instance name writes and queries will be run against.")
	f.StringVar(&cfg.WorkloadFilePath, "bench.workload-file-path", "./workload.yaml", "path to the file containing the workload description")

	cfg.Write.RegisterFlags(f)
	cfg.RingCheck.RegisterFlagsWithPrefix("bench.ring-check.", f)
}

type BenchRunner struct {
	cfg Config

	writeRunner     *WriteBenchmarkRunner
	ringCheckRunner *RingChecker
}

func NewBenchRunner(cfg Config, logger log.Logger, reg prometheus.Registerer) (*BenchRunner, error) {
	// Load workload file

	content, err := os.ReadFile(cfg.WorkloadFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read workload YAML file from the disk")
	}

	workloadDesc := WorkloadDesc{}
	err = yaml.Unmarshal(content, &workloadDesc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal workload YAML file")
	}

	level.Info(logger).Log("msg", "building workload")
	workload := newWorkload(workloadDesc, prometheus.DefaultRegisterer)

	benchRunner := &BenchRunner{
		cfg: cfg,
	}

	if cfg.Write.Enabled {
		benchRunner.writeRunner, err = NewWriteBenchmarkRunner(cfg.ID, cfg.Write, workload, logger, reg)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create write benchmarker")
		}
	}

	if cfg.RingCheck.Enabled {
		benchRunner.ringCheckRunner, err = NewRingChecker(cfg.ID, cfg.InstanceName, cfg.RingCheck, workload, logger)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create ring checker")
		}
	}
	return benchRunner, nil
}

func (b *BenchRunner) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	if b.writeRunner != nil {
		g.Go(func() error {
			return b.writeRunner.Run(ctx)
		})
	}

	if b.ringCheckRunner != nil {
		g.Go(func() error {
			return b.ringCheckRunner.Run(ctx)
		})
	}

	return g.Wait()
}

type WriteBenchConfig struct {
	Enabled           bool   `yaml:"enabled"`
	Endpoint          string `yaml:"endpoint"`
	BasicAuthUsername string `yaml:"basic_auth_username"`
	BasicAuthPasword  string `yaml:"basic_auth_password"`

	Interval  time.Duration `yaml:"interval"`
	Timeout   time.Duration `yaml:"timeout"`
	BatchSize int           `yaml:"batch_size"`
}

func (cfg *WriteBenchConfig) RegisterFlags(f *flag.FlagSet) {
	f.BoolVar(&cfg.Enabled, "bench.write.enabled", true, "enable write benchmarking")
	f.StringVar(&cfg.Endpoint, "bench.write.endpoint", "", "Remote write endpoint.")
	f.StringVar(&cfg.BasicAuthUsername, "bench.write.basic-auth-username", "", "Set the basic auth username on remote write requests.")
	f.StringVar(&cfg.BasicAuthPasword, "bench.write.basic-auth-password", "", "Set the basic auth password on remote write requests.")

	f.DurationVar(&cfg.Interval, "bench.write.interval", time.Second*15, "Interval between sending each batch of series.")
	f.DurationVar(&cfg.Timeout, "bench.write.timeout", time.Second*30, "Write timeout for sending remote write series.")
	f.IntVar(&cfg.BatchSize, "bench.write.batch-size", 500, "Number of samples to send per remote-write request")
}

type WriteBenchmarkRunner struct {
	id  string
	cfg WriteBenchConfig

	// Do DNS client side load balancing if configured
	remoteMtx  sync.Mutex
	addresses  []string
	clientPool map[string]*writeClient

	dnsProvider *dns.Provider

	workload *workload

	reg    prometheus.Registerer
	logger log.Logger

	requestDuration *prometheus.HistogramVec
}

func NewWriteBenchmarkRunner(id string, cfg WriteBenchConfig, workload *workload, logger log.Logger, reg prometheus.Registerer) (*WriteBenchmarkRunner, error) {
	writeBench := &WriteBenchmarkRunner{
		id:  id,
		cfg: cfg,

		workload: workload,
		dnsProvider: dns.NewProvider(
			logger,
			extprom.WrapRegistererWithPrefix("benchtool_", reg),
			dns.GolangResolverType,
		),
		clientPool: map[string]*writeClient{},
		logger:     logger,
		reg:        reg,
		requestDuration: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "benchtool",
				Name:      "write_request_duration_seconds",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120},
			},
			[]string{"code"},
		),
	}

	// Resolve an initial set of distributor addresses
	err := writeBench.resolveAddrs()
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve enpoints")
	}

	return writeBench, nil
}

func (w *WriteBenchmarkRunner) getRandomWriteClient() (*writeClient, error) {
	w.remoteMtx.Lock()
	defer w.remoteMtx.Unlock()

	randomIndex := rand.Intn(len(w.addresses))
	pick := w.addresses[randomIndex]

	var cli *writeClient
	var exists bool

	if cli, exists = w.clientPool[pick]; !exists {
		u, err := url.Parse("http://" + pick + "/api/v1/push")
		if err != nil {
			return nil, err
		}
		cli, err = newWriteClient("bench-"+pick, &remote.ClientConfig{
			URL:     &config.URL{URL: u},
			Timeout: model.Duration(w.cfg.Timeout),

			HTTPClientConfig: config.HTTPClientConfig{
				BasicAuth: &config.BasicAuth{
					Username: w.cfg.BasicAuthUsername,
					Password: config.Secret(w.cfg.BasicAuthPasword),
				},
			},
		}, w.requestDuration)
		if err != nil {
			return nil, err
		}
		w.clientPool[pick] = cli
	}

	return cli, nil
}

func (w *WriteBenchmarkRunner) Run(ctx context.Context) error {
	// Start a loop to re-resolve addresses every 5 minutes
	go w.resolveAddrsLoop(ctx)

	batchChan := make(chan []prompb.TimeSeries, 10)
	for i := 0; i < 10; i++ {
		level.Info(w.logger).Log("msg", "starting worker", "worker_num", strconv.Itoa(i))
		go w.worker(batchChan)
	}

	ticker := time.NewTicker(w.cfg.Interval)
	for {
		select {
		case <-ctx.Done():
			close(batchChan)
			return nil
		case <-ticker.C:
			timeseries := w.workload.generateTimeSeries(w.id)
			batchSize := w.cfg.BatchSize
			var batches [][]prompb.TimeSeries
			if batchSize < len(timeseries) {
				batches = make([][]prompb.TimeSeries, 0, (len(timeseries)+batchSize-1)/batchSize)

				level.Info(w.logger).Log("msg", "sending timeseries", "num_series", strconv.Itoa(len(timeseries)))
				for batchSize < len(timeseries) {
					timeseries, batches = timeseries[batchSize:], append(batches, timeseries[0:batchSize:batchSize])
				}
			} else {
				batches = [][]prompb.TimeSeries{timeseries}
			}

			for _, batch := range batches {
				batchChan <- batch
			}
		}
	}
}

func (w *WriteBenchmarkRunner) worker(batchChannel chan []prompb.TimeSeries) {
	for batch := range batchChannel {
		err := w.sendBatch(batch)
		if err != nil {
			level.Error(w.logger).Log("msg", "failed to send batch", "err", err)
		}
	}
}

func (w *WriteBenchmarkRunner) sendBatch(batch []prompb.TimeSeries) error {
	level.Debug(w.logger).Log("msg", "sending timeseries batch", "num_series", strconv.Itoa(len(batch)))
	cli, err := w.getRandomWriteClient()
	if err != nil {
		return errors.Wrap(err, "unable to get remote-write client")
	}
	req := prompb.WriteRequest{
		Timeseries: batch,
	}

	data, err := proto.Marshal(&req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal remote-write request")
	}

	compressed := snappy.Encode(nil, data)

	err = cli.Store(context.Background(), compressed)

	if err != nil {
		return errors.Wrap(err, "remote-write request failed")
	}

	return nil
}

func (w *WriteBenchmarkRunner) resolveAddrsLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := w.resolveAddrs()
			if err != nil {
				level.Warn(w.logger).Log("msg", "failed update remote write servers list", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (w *WriteBenchmarkRunner) resolveAddrs() error {
	// Resolve configured addresses with a reasonable timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// If some of the dns resolution fails, log the error.
	if err := w.dnsProvider.Resolve(ctx, []string{w.cfg.Endpoint}); err != nil {
		level.Error(w.logger).Log("msg", "failed to resolve addresses", "err", err)
	}

	// Fail in case no server address is resolved.
	servers := w.dnsProvider.Addresses()
	if len(servers) == 0 {
		return errors.New("no server address resolved")
	}

	w.remoteMtx.Lock()
	w.addresses = servers
	w.remoteMtx.Unlock()

	return nil
}

type RingCheckConfig struct {
	Enabled       bool                `yaml:"enabled"`
	MemberlistKV  memberlist.KVConfig `yaml:"memberlist"`
	RingConfig    ring.Config         `yaml:"ring"`
	CheckInterval time.Duration       `yaml:"check_interval"`
}

func (cfg *RingCheckConfig) RegisterFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.BoolVar(&cfg.Enabled, prefix+"enabled", true, "enable ring check module")
	cfg.MemberlistKV.RegisterFlags(f, prefix)
	cfg.RingConfig.RegisterFlagsWithPrefix(prefix, f)

	f.DurationVar(&cfg.CheckInterval, prefix+"check-interval", 5*time.Minute, "Interval at which the current ring will be compared with the configured workload")
}

type RingChecker struct {
	id           string
	instanceName string
	cfg          RingCheckConfig

	Ring         *ring.Ring
	MemberlistKV *memberlist.KVInitService
	workload     *workload
	logger       log.Logger
}

func NewRingChecker(id string, instanceName string, cfg RingCheckConfig, workload *workload, logger log.Logger) (*RingChecker, error) {
	r := RingChecker{
		id:           id,
		instanceName: instanceName,
		cfg:          cfg,

		logger:   logger,
		workload: workload,
	}
	cfg.MemberlistKV.MetricsRegisterer = prometheus.DefaultRegisterer
	cfg.MemberlistKV.Codecs = []codec.Codec{
		ring.GetCodec(),
	}
	r.MemberlistKV = memberlist.NewKVInitService(&cfg.MemberlistKV, logger)
	cfg.RingConfig.KVStore.MemberlistKV = r.MemberlistKV.GetMemberlistKV

	var err error
	r.Ring, err = ring.New(cfg.RingConfig, "ingester", ring.IngesterRingKey, prometheus.DefaultRegisterer)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *RingChecker) Run(ctx context.Context) error {
	err := r.Ring.Service.StartAsync(ctx)
	if err != nil {
		return fmt.Errorf("unable to start ring, %w", err)
	}
	ticker := time.NewTicker(r.cfg.CheckInterval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			r.check()
		}
	}
}

func (r *RingChecker) check() {
	timeseries := r.workload.generateTimeSeries(r.id)

	addrMap := map[string]int{}
	for _, s := range timeseries {
		sort.Slice(s.Labels, func(i, j int) bool {
			return strings.Compare(s.Labels[i].Name, s.Labels[j].Name) < 0
		})

		token := shardByAllLabels(r.instanceName, s.Labels)

		rs, err := r.Ring.Get(token, ring.Write, []ring.IngesterDesc{})

		if err != nil {
			level.Warn(r.logger).Log("msg", "unable to get token for metric", "err", err)
			continue
		}

		rs.GetAddresses()
		for _, addr := range rs.GetAddresses() {
			_, exists := addrMap[addr]
			if !exists {
				addrMap[addr] = 0
			}
			addrMap[addr] += 1
		}
	}

	fmt.Println("ring check:")
	for addr, tokensTotal := range addrMap {
		fmt.Printf("  %s,%d\n", addr, tokensTotal)
	}
}

func shardByUser(userID string) uint32 {
	h := ingester_client.HashNew32()
	h = ingester_client.HashAdd32(h, userID)
	return h
}

// This function generates different values for different order of same labels.
func shardByAllLabels(userID string, labels []prompb.Label) uint32 {
	h := shardByUser(userID)
	for _, label := range labels {
		h = ingester_client.HashAdd32(h, label.Name)
		h = ingester_client.HashAdd32(h, label.Value)
	}
	return h
}
