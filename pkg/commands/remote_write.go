package commands

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type RemoteWriteOOOCommand struct {
	address         string
	remoteWritePath string

	tenantID string
	apiKey   string

	metricName string

	batchSize     int
	writeInterval time.Duration

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

	remoteWriteCmd.Flag("timeout", "timeout for write requests").
		Default("30s").
		DurationVar(&c.timeout)

	remoteWriteCmd.Flag("batch-size", "how many samples get written with each request").
		Default("1000").
		IntVar(&c.batchSize)

	remoteWriteCmd.Flag("write-interval", "interval at which batches are written").
		Default("1s").
		DurationVar(&c.writeInterval)
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
	writeClient, err := c.writeClient()
	if err != nil {
		return err
	}

	Labels := []prompb.Label{
		{Name: "__name__", Value: c.metricName},
		{Name: "job", Value: "node_exporter"},
		{Name: "instance", Value: "test_instance"},
		{Name: "cpu", Value: "0"},
		{Name: "mode", Value: "idle"},
	}
	req := prompb.WriteRequest{
		Timeseries: make([]prompb.TimeSeries, 0, c.batchSize),
	}
	ticker := time.NewTicker(c.writeInterval)
	for range ticker.C {
		var ts int64
		for i := int64(0); i < int64(c.batchSize); i++ {
			if i == 0 {
				ts = time.Now().Unix() * 1000
			} else if i == 1 {
				// All except the first sample are out of order, because they're 1 second older than the first sample of the batch
				ts = ts - 1000
			}

			req.Timeseries = append(req.Timeseries, prompb.TimeSeries{
				Labels: Labels,
				Samples: []prompb.Sample{
					{
						Timestamp: ts,
						Value:     float64(i),
					},
				},
			})
		}

		data, err := proto.Marshal(&req)
		if err != nil {
			return err
		}

		err = writeClient.Store(context.Background(), snappy.Encode(nil, data))
		if err != nil {
			fmt.Printf("Error writing: %s", err)
		}
	}

	return nil
}
