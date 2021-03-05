package bench

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	"github.com/thanos-io/thanos/pkg/discovery/dns"
	"github.com/thanos-io/thanos/pkg/extprom"
	"gopkg.in/yaml.v2"
)

type SeriesType string

const (
	GaugeZero     SeriesType = "gauge-zero"
	GaugeRandom   SeriesType = "gauge-random"
	CounterOne    SeriesType = "counter-one"
	CounterRandom SeriesType = "counter-random"
)

type LabelDesc struct {
	Name         string `yaml:"name"`
	ValuePrefix  string `yaml:"value_prefix"`
	UniqueValues int    `yaml:"unique_values"`
}

type SeriesDesc struct {
	Name         string            `yaml:"name"`
	Type         SeriesType        `yaml:"type"`
	StaticLabels map[string]string `yaml:"static_labels"`
	Labels       []LabelDesc       `yaml:"labels"`
}

type WorkloadDesc struct {
	Replicas int          `yaml:"replicas"`
	Series   []SeriesDesc `yaml:"series"`
}

type timeseries struct {
	labelSets  [][]prompb.Label
	lastValue  float64
	seriesType SeriesType
}

type workload struct {
	replicas    int
	series      []*timeseries
	totalSeries int
}

func newWorkload(workloadDesc WorkloadDesc) *workload {
	totalSeries := 0
	series := []*timeseries{}

	for _, seriesDesc := range workloadDesc.Series {
		// Create the metric with a name value
		labelSets := [][]prompb.Label{
			{
				prompb.Label{Name: "__name__", Value: seriesDesc.Name},
			},
		}

		// Add any configured static labels
		for labelName, labelValue := range seriesDesc.StaticLabels {
			labelSets[0] = append(labelSets[0], prompb.Label{Name: labelName, Value: labelValue})
		}

		// Create the dynamic label set
		for _, lbl := range seriesDesc.Labels {
			newLabelSets := make([][]prompb.Label, 0, len(labelSets)*lbl.UniqueValues)
			for i := 0; i < lbl.UniqueValues; i++ {
				for _, labelSet := range labelSets {
					newLabelSet := append(labelSet, prompb.Label{
						Name:  lbl.Name,
						Value: fmt.Sprintf("%s-%v", lbl.ValuePrefix, i),
					},
					)
					newLabelSets = append(newLabelSets, newLabelSet)
				}
			}
			labelSets = newLabelSets
		}

		series = append(series, &timeseries{
			labelSets:  labelSets,
			seriesType: seriesDesc.Type,
		})
		totalSeries += len(labelSets)
	}

	return &workload{
		replicas:    workloadDesc.Replicas,
		series:      series,
		totalSeries: totalSeries,
	}
}

func (w *workload) generateTimeSeries(id string) []prompb.TimeSeries {
	now := time.Now().UnixNano() / int64(time.Millisecond)

	timeseries := make([]prompb.TimeSeries, 0, w.replicas*w.totalSeries)
	for replicaNum := 0; replicaNum < w.replicas; replicaNum++ {
		replicaLabel := prompb.Label{Name: "bench_replica", Value: fmt.Sprintf("%s-replica-%05d", id, replicaNum)}
		for _, series := range w.series {
			var value float64
			switch series.seriesType {
			case GaugeZero:
				value = 0
			case GaugeRandom:
				value = rand.Float64()
			case CounterOne:
				value = series.lastValue + 1
			case CounterRandom:
				value = series.lastValue + float64(rand.Int())
			default:
				panic(fmt.Sprintf("unknown series type %v", series.seriesType))
			}
			series.lastValue = value
			for _, labelSet := range series.labelSets {
				timeseries = append(timeseries, prompb.TimeSeries{
					Labels: append(labelSet, replicaLabel),
					Samples: []prompb.Sample{{
						Timestamp: now,
						Value:     value,
					}},
				})
			}
		}
	}

	return timeseries
}

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
	clientPool map[string]remote.WriteClient

	dnsProvider *dns.Provider

	workload *workload

	logger log.Logger
}

func NewWriteBench(cfg WriteBenchConfig, logger log.Logger, reg prometheus.Registerer) (*WriteBench, error) {
	// Load workload file

	content, err := ioutil.ReadFile(cfg.WorkloadFilePath)
	if err != nil {
		return nil, err
	}

	workloadDesc := WorkloadDesc{}
	err = yaml.Unmarshal(content, &workloadDesc)
	if err != nil {
		return nil, err
	}

	workload := newWorkload(workloadDesc)
	if err != nil {
		return nil, err
	}

	writeBench := &WriteBench{
		cfg: cfg,

		workload: workload,
		dnsProvider: dns.NewProvider(
			logger,
			extprom.WrapRegistererWithPrefix("benchtool_", reg),
			dns.GolangResolverType,
		),
		logger: logger,
	}

	// Resolve an initial set of distributor addresses
	err = writeBench.resolveAddrs()
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve enpoints")
	}

	return writeBench, nil
}

func (w *WriteBench) getRandomWriteClient() (remote.WriteClient, error) {
	w.remoteMtx.Lock()
	defer w.remoteMtx.Unlock()

	randomIndex := rand.Intn(len(w.addresses))
	pick := w.addresses[randomIndex]

	var cli remote.WriteClient
	var exists bool

	if cli, exists = w.clientPool[pick]; !exists {
		u, err := url.Parse("http://" + pick + "/api/v1/push")
		if err != nil {
			return nil, err
		}
		cli, err = remote.NewWriteClient("bench-"+pick, &remote.ClientConfig{
			URL:     &config.URL{URL: u},
			Timeout: model.Duration(w.cfg.WriteTimeout),

			HTTPClientConfig: config.HTTPClientConfig{
				BasicAuth: &config.BasicAuth{
					Username: w.cfg.BasicAuthUsername,
					Password: config.Secret(w.cfg.BasicAuthPasword),
				},
			},
		})
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
		go w.worker(batchChan)
	}

	ticker := time.NewTicker(w.cfg.SendInterval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			timeseries := w.workload.generateTimeSeries(w.cfg.ID)
			batchSize := w.cfg.BatchSize
			batches := make([][]prompb.TimeSeries, 0, (len(timeseries)+batchSize-1)/batchSize)

			for batchSize < len(timeseries) {
				timeseries, batches = timeseries[batchSize:], append(batches, timeseries[0:batchSize:batchSize])
			}

			for _, batch := range batches {
				batchChan <- batch
			}
		}
	}
}

func (w *WriteBench) worker(batchChannel chan []prompb.TimeSeries) {
	for batch := range batchChannel {
		cli, err := w.getRandomWriteClient()
		if err != nil {
			level.Error(w.logger).Log("msg", "unable to get client", "err", err)
			continue
		}
		req := prompb.WriteRequest{
			Timeseries: batch,
		}

		data, err := proto.Marshal(&req)
		if err != nil {
			level.Error(w.logger).Log("msg", "unable to marshal write request", "err", err)
			continue
		}

		compressed := snappy.Encode(nil, data)

		if err := cli.Store(context.Background(), compressed); err != nil {
			level.Error(w.logger).Log("msg", "unable to write request", "err", err)
			continue
		}
	}
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
