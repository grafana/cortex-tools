package bench

import (
	"context"
	"flag"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

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
	"gopkg.in/yaml.v2"
)

type WriteBenchConfig struct {
	ID                string `yaml:"id"`
	Endpoint          string `yaml:"endpoint"`
	HeaderID          string `yaml:"header_id"`
	BasicAuthUsername string `yaml:"basic_auth_username"`
	BasicAuthPasword  string `yaml:"basic_auth_password"`

	SendInterval time.Duration `yaml:"send_interval"`
	WriteTimeout time.Duration `yaml:"write_timeout"`

	BatchSize        int    `yaml:"batch_size"`
	WorkloadFilePath string `yaml:"workload_file_path"`
}

func (cfg *WriteBenchConfig) RegisterFlags(f *flag.FlagSet) {
	defaultID, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	f.StringVar(&cfg.ID, "write-bench.id", defaultID, "ID of worker. Defaults to hostname")
	f.StringVar(&cfg.Endpoint, "write-bench.endpoint", "", "Remote write endpoint.")
	f.StringVar(&cfg.HeaderID, "write-bench.header-id", "", "Sets the X-Scope-OrgID header on write requests to this value.")
	f.StringVar(&cfg.BasicAuthUsername, "write-bench.basic-auth-username", "", "Set the basic auth username on remote write requests.")
	f.StringVar(&cfg.BasicAuthPasword, "write-bench.basic-auth-password", "", "Set the basic auth password on remote write requests.")

	f.DurationVar(&cfg.SendInterval, "write-bench.send-interval", time.Second*15, "Interval between sending each batch of series.")
	f.DurationVar(&cfg.WriteTimeout, "write-bench.write-timeout", time.Second*30, "Write timeout for sending remote write series.")
	f.IntVar(&cfg.BatchSize, "write-bench.batch-size", 500, "Number of samples to send per remote-write request")

	f.StringVar(&cfg.WorkloadFilePath, "write-bench.workload-file-path", "./workload.yaml", "path to the file containing the workload description")
}

type WriteBench struct {
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

func NewWriteBench(cfg WriteBenchConfig, logger log.Logger, reg prometheus.Registerer) (*WriteBench, error) {
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
	workload := newWorkload(workloadDesc)

	writeBench := &WriteBench{
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
	err = writeBench.resolveAddrs()
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve enpoints")
	}

	return writeBench, nil
}

func (w *WriteBench) getRandomWriteClient() (*writeClient, error) {
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
			Timeout: model.Duration(w.cfg.WriteTimeout),

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

func (w *WriteBench) Run(ctx context.Context) error {
	// Start a loop to re-resolve addresses every 5 minutes
	go w.resolveAddrsLoop(ctx)

	batchChan := make(chan []prompb.TimeSeries, 10)
	for i := 0; i < 10; i++ {
		level.Info(w.logger).Log("msg", "starting worker", "worker_num", strconv.Itoa(i))
		go w.worker(batchChan)
	}

	ticker := time.NewTicker(w.cfg.SendInterval)
	for {
		select {
		case <-ctx.Done():
			close(batchChan)
			return nil
		case <-ticker.C:
			timeseries := w.workload.generateTimeSeries(w.cfg.ID)
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

func (w *WriteBench) worker(batchChannel chan []prompb.TimeSeries) {
	for batch := range batchChannel {
		err := w.sendBatch(batch)
		if err != nil {
			level.Error(w.logger).Log("msg", "failed to send batch", "err", err)
		}
	}
}

func (w *WriteBench) sendBatch(batch []prompb.TimeSeries) error {

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

func (w *WriteBench) resolveAddrsLoop(ctx context.Context) {
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

func (w *WriteBench) resolveAddrs() error {
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
