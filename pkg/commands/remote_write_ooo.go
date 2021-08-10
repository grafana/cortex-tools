package commands

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"
)

type RemoteWriteOOOCommand struct {
	address         string
	remoteWritePath string

	tenantID string
	apiKey   string

	metricName  string
	seriesCount int

	threadCount   int
	batchSize     int
	writeInterval time.Duration
	verbose       bool

	timeout time.Duration
}

func (c *RemoteWriteOOOCommand) Register(app *kingpin.Application) {
	remoteWriteCmd := app.Command("remote-write-ooo", "Remote Write to Cortex.").Action(c.remoteWriteOOO)

	remoteWriteCmd.Flag("address", "Address of the cortex cluster, alternatively set $CORTEX_ADDRESS.").
		Envar("CORTEX_ADDRESS").
		Required().
		StringVar(&c.address)

	remoteWriteCmd.Flag("remote-write-path", "Path of the remote read endpoint.").
		Default("/api/prom/push").
		StringVar(&c.remoteWritePath)

	remoteWriteCmd.Flag("id", "Cortex tenant id, alternatively set $CORTEX_TENANT_ID.").
		Envar("CORTEX_TENANT_ID").
		Default("").
		StringVar(&c.tenantID)

	remoteWriteCmd.Flag("key", "Api key to use when contacting cortex, alternatively set $CORTEX_API_KEY.").
		Envar("CORTEX_API_KEY").
		Default("").
		StringVar(&c.apiKey)

	remoteWriteCmd.Flag("metric-name", "Name of the test metric to write.").
		Default("test_metric").
		StringVar(&c.metricName)

	remoteWriteCmd.Flag("series-count", "Number of series to write.").
		Default("10").
		IntVar(&c.seriesCount)

	remoteWriteCmd.Flag("thread-count", "Number of threads that write concurrently.").
		Default("10").
		IntVar(&c.threadCount)

	remoteWriteCmd.Flag("timeout", "timeout for write requests").
		Default("30s").
		DurationVar(&c.timeout)

	remoteWriteCmd.Flag("batch-size", "how many samples get written per series with each request").
		Default("100").
		IntVar(&c.batchSize)

	remoteWriteCmd.Flag("write-interval", "interval at which batches are written").
		Default("5s").
		DurationVar(&c.writeInterval)

	remoteWriteCmd.Flag("verbose", "write all samples that get sent").
		Default("false").
		BoolVar(&c.verbose)
}

func (c *RemoteWriteOOOCommand) writeClient() (remote.WriteClient, error) {
	addressURL, err := url.Parse(c.address)
	if err != nil {
		return nil, err
	}

	addressURL.Path = filepath.Join(
		addressURL.Path,
		c.remoteWritePath,
	)

	writeClient, err := remote.NewWriteClient("remote-write", &remote.ClientConfig{
		URL:     &config_util.URL{URL: addressURL},
		Timeout: model.Duration(c.timeout),
		HTTPClientConfig: config_util.HTTPClientConfig{
			BasicAuth: &config_util.BasicAuth{
				Username: c.tenantID,
				Password: config_util.Secret(c.apiKey),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if c.tenantID != "" {
		client, ok := writeClient.(*remote.Client)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T", writeClient)
		}
		client.Client.Transport = &setTenantIDTransport{
			RoundTripper: client.Client.Transport,
			tenantID:     c.tenantID,
		}
	}

	log.Infof("Created remote write client using endpoint '%s'", redactedURL(addressURL))

	return writeClient, nil
}

func (c *RemoteWriteOOOCommand) remoteWriteOOO(k *kingpin.ParseContext) error {
	labels := []prompb.Label{
		{Name: "__name__", Value: c.metricName},
		{Name: "job", Value: "node_exporter"},
		{Name: "instance", Value: "test_instance"},
		{Name: "cpu", Value: "0"},
		{Name: "mode", Value: "idle"},
	}
	sort.Slice(labels, func(i, j int) bool {
		return strings.Compare(labels[i].Name, labels[j].Name) < 0
	})

	requestCh := c.startWorkers()
	ticker := time.NewTicker(c.writeInterval)
	samples := make([]prompb.Sample, 0, c.batchSize)
	var ts int64
	for range ticker.C {
		samples = samples[:0]
		for sampleIdx := 0; sampleIdx < c.batchSize; sampleIdx++ {
			if sampleIdx == 0 {
				ts = time.Now().Unix() * 1000
			} else if sampleIdx == 1 {
				// All except the first sample of each series are out of order
				// because they're older than the first sample of the batch
				ts = ts - 1000 - c.writeInterval.Milliseconds()
			}

			samples = append(samples, prompb.Sample{
				Timestamp: ts,
				Value:     float64(sampleIdx),
			})
		}

		for seriesIdx := 0; seriesIdx < c.seriesCount; seriesIdx++ {
			seriesLabel := prompb.Label{
				Name:  "unique_label",
				Value: strconv.Itoa(seriesIdx),
			}
			seriesLabels := append(labels, seriesLabel)

			if c.verbose {
				var seriesString string
				for labelIdx, label := range seriesLabels {
					if labelIdx > 0 {
						seriesString = seriesString + ";"
					}
					seriesString = seriesString + label.Name + "=" + label.Value
				}

				fmt.Printf("Series: %s\n", seriesString)

				for _, sample := range samples {
					fmt.Printf("ts: %d, value: %f, ", sample.Timestamp, sample.Value)
				}
				fmt.Printf("\n")
			}

			data, err := proto.Marshal(&prompb.WriteRequest{
				Timeseries: []prompb.TimeSeries{{
					Labels:  seriesLabels,
					Samples: samples,
				}},
			})
			if err != nil {
				fmt.Printf("failed to marshal request: %s\n", err)
				return err
			}

			requestCh <- data
		}
	}

	return nil
}

func (c *RemoteWriteOOOCommand) startWorkers() chan []byte {
	requestCh := make(chan []byte, c.seriesCount)

	worker := func() error {
		writeClient, err := c.writeClient()
		if err != nil {
			fmt.Printf("failed to instantiate client: %s\n", err)
			return err
		}

		for req := range requestCh {
			err = writeClient.Store(context.Background(), snappy.Encode(nil, req))
			if err != nil {
				fmt.Printf("Error writing request: %s", err)
			}
		}

		return nil
	}

	group, _ := errgroup.WithContext(context.Background())
	for i := 0; i < c.threadCount; i++ {
		group.Go(worker)
	}

	return requestCh
}
