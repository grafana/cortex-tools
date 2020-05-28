package commands

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
	"gopkg.in/alecthomas/kingpin.v2"
)

var writeRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "write_requests_duration_seconds",
	Buckets: prometheus.DefBuckets,
}, []string{"success"})

type LoadgenCommand struct {
	url            string
	activeSeries   int
	scrapeInterval time.Duration
	parallelism    int
	batchSize      int
	timeout        time.Duration

	metricsListenAddress string

	wg     sync.WaitGroup
	client *remote.Client
}

func (c *LoadgenCommand) Register(app *kingpin.Application) {
	loadgenCommand := &LoadgenCommand{}
	cmd := app.Command("loadgen", "Simple load generator for Cortex.").Action(loadgenCommand.run)
	cmd.Flag("url", "").
		Required().StringVar(&loadgenCommand.url)
	cmd.Flag("active-series", "number of active series to send").
		Default("1000").IntVar(&loadgenCommand.activeSeries)
	cmd.Flag("scrape-interval", "period to send metrics").
		Default("15s").DurationVar(&loadgenCommand.scrapeInterval)
	cmd.Flag("parallelism", "how many metrics to send simultaneously").
		Default("10").IntVar(&loadgenCommand.parallelism)
	cmd.Flag("batch-size", "how big a batch to send").
		Default("100").IntVar(&loadgenCommand.batchSize)
	cmd.Flag("request-timeout", "timeout for write requests").
		Default("500ms").DurationVar(&loadgenCommand.timeout)
	cmd.Flag("metrics-listen-address", "address to serve metrics on").
		Default(":8080").StringVar(&loadgenCommand.metricsListenAddress)
}

func (c *LoadgenCommand) run(k *kingpin.ParseContext) error {
	url, err := url.Parse(c.url)
	if err != nil {
		return err
	}

	client, err := remote.NewClient(0, &remote.ClientConfig{
		URL:     &config.URL{URL: url},
		Timeout: model.Duration(c.timeout),
	})
	if err != nil {
		return err
	}
	c.client = client

	http.Handle("/metrics", promhttp.Handler())
	go log.Fatal(http.ListenAndServe(c.metricsListenAddress, nil))

	c.wg.Add(c.parallelism)

	metricsPerShard := c.activeSeries / c.parallelism
	for i := 0; i < c.activeSeries; i += metricsPerShard {
		go c.runShard(i, i+metricsPerShard)
	}

	c.wg.Wait()
	return nil
}

func (c *LoadgenCommand) runShard(from, to int) {
	defer c.wg.Done()
	ticker := time.NewTicker(c.scrapeInterval)
	c.runScrape(from, to)
	for range ticker.C {
		c.runScrape(from, to)
	}
}

func (c *LoadgenCommand) runScrape(from, to int) {
	for i := from; i < to; i += c.batchSize {
		if err := c.runBatch(i, i+c.batchSize); err != nil {
			log.Printf("error sending batch: %v", err)
		}
	}
	fmt.Printf("sent %d samples\n", to-from)
}

func (c *LoadgenCommand) runBatch(from, to int) error {
	var (
		req = prompb.WriteRequest{
			Timeseries: make([]prompb.TimeSeries, 0, to-from),
		}
		now = time.Now().UnixNano() / int64(time.Millisecond)
	)

	for i := from; i < to; i++ {
		timeseries := prompb.TimeSeries{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "node_cpu_seconds_total"},
				{Name: "job", Value: "node_exporter"},
				{Name: "instance", Value: fmt.Sprintf("instance%000d", i)},
				{Name: "cpu", Value: "0"},
				{Name: "mode", Value: "idle"},
			},
			Samples: []prompb.Sample{{
				Timestamp: now,
				Value:     rand.Float64(),
			}},
		}
		req.Timeseries = append(req.Timeseries, timeseries)
	}

	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}

	compressed := snappy.Encode(nil, data)

	start := time.Now()
	if err := c.client.Store(context.Background(), compressed); err != nil {
		writeRequestDuration.WithLabelValues("error").Observe(time.Now().Sub(start).Seconds())
		return err
	}
	writeRequestDuration.WithLabelValues("success").Observe(time.Now().Sub(start).Seconds())

	return nil
}
