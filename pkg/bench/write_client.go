package bench

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cortexproject/cortex/pkg/util/spanlogger"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/version"
	"github.com/prometheus/prometheus/storage/remote"
)

const maxErrMsgLen = 512

var UserAgent = fmt.Sprintf("Prometheus/%s", version.Version)

// writeClient allows reading and writing from/to a remote HTTP endpoint.
type writeClient struct {
	remoteName string // Used to differentiate clients in metrics.
	url        *config_util.URL
	Client     *http.Client
	timeout    time.Duration

	requestDuration *prometheus.HistogramVec
}

// newWriteClient creates a new client for remote write.
func newWriteClient(name string, conf *remote.ClientConfig, reg prometheus.Registerer) (*writeClient, error) {
	httpClient, err := config_util.NewClientFromConfig(conf.HTTPClientConfig, "remote_storage_write_client", false, false)
	if err != nil {
		return nil, err
	}

	t := httpClient.Transport
	httpClient.Transport = &nethttp.Transport{
		RoundTripper: t,
	}

	return &writeClient{
		remoteName: name,
		url:        conf.URL,
		Client:     httpClient,
		timeout:    time.Duration(conf.Timeout),

		requestDuration: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "benchtool",
				Name:      "write_request_duration_seconds",
				Buckets:   []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120},
			},
			[]string{"code"},
		),
	}, nil
}

type RecoverableError struct {
	error
}

// Store sends a batch of samples to the HTTP endpoint, the request is the proto marshalled
// and encoded bytes from codec.go.
func (c *writeClient) Store(ctx context.Context, req []byte) error {
	spanLog, ctx := spanlogger.New(ctx, "writeClient.Store")
	defer spanLog.Span.Finish()
	httpReq, err := http.NewRequest("POST", c.url.String(), bytes.NewReader(req))
	if err != nil {
		// Errors from NewRequest are from unparsable URLs, so are not
		// recoverable.
		return err
	}
	httpReq.Header.Add("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("User-Agent", UserAgent)
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	httpReq = httpReq.WithContext(ctx)

	start := time.Now()

	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		// Errors from Client.Do are from (for example) network errors, so are
		// recoverable.
		return RecoverableError{err}
	}
	c.requestDuration.WithLabelValues(httpResp.Status).Observe(time.Since(start).Seconds())

	defer func() {
		io.Copy(ioutil.Discard, httpResp.Body)
		httpResp.Body.Close()
	}()

	if httpResp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(httpResp.Body, maxErrMsgLen))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = errors.Errorf("server returned HTTP status %s: %s", httpResp.Status, line)
	}
	if httpResp.StatusCode/100 == 5 {
		return RecoverableError{err}
	}
	return err
}
