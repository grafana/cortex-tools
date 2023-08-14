package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func Test_healthCheck(t *testing.T) {
	r, err := NewReceiver(
		ReceiverConfig{PurgeInterval: 1 * time.Hour},
		log.NewNopLogger(),
		prometheus.NewRegistry(),
	)
	require.NoError(t, err)

	router := mux.NewRouter()
	r.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func Test_measureLatency(t *testing.T) {
	tc := []struct {
		name    string
		alerts  template.Data
		err     error
		tracked []float64
	}{
		{
			name: "with alerts to track",
			alerts: template.Data{
				Alerts: template.Alerts{
					template.Alert{
						Labels:      template.KV{model.AlertNameLabel: "e2ealertingAlwaysFiring"},
						Annotations: template.KV{"time": "1.604069614e+09"},
						Status:      string(model.AlertFiring),
					},
					template.Alert{
						Labels:      template.KV{model.AlertNameLabel: "e2ealertingAlwaysFiring"},
						Annotations: template.KV{"time": "1.604069615e+09"},
						Status:      string(model.AlertFiring),
					},
				},
			},
			tracked: []float64{1604069614.00, 1604069615.00},
		},
		{
			name: "with alerts that don't have a time annotation or alertname label it ignores them",
			alerts: template.Data{
				Alerts: template.Alerts{
					template.Alert{
						Labels:      template.KV{model.AlertNameLabel: "e2ealertingAlwaysFiring"},
						Annotations: template.KV{"time": "1.604069614e+09"},
						Status:      string(model.AlertFiring),
					},
					template.Alert{
						Labels: template.KV{model.AlertNameLabel: "e2ealertingAlwaysFiring"},
						Status: string(model.AlertFiring),
					},
					template.Alert{
						Annotations: template.KV{"time": "1.604069614e+09"},
						Status:      string(model.AlertFiring),
					},
				},
			},
			tracked: []float64{1604069614.00},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewReceiver(
				ReceiverConfig{PurgeInterval: 1 * time.Hour},
				log.NewNopLogger(),
				prometheus.NewRegistry(),
			)
			require.NoError(t, err)

			router := mux.NewRouter()
			r.RegisterRoutes(router)

			b, err := json.Marshal(tt.alerts)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/receiver", bytes.NewBuffer(b))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, len(tt.tracked), len(r.timestamps))
			for _, timestamp := range tt.tracked {
				_, exists := r.timestamps[timestamp]
				require.True(t, exists, fmt.Sprintf("time %f is not tracked", timestamp))
			}
		})
	}
}

func Test_measureLatencyCustomBuckets(t *testing.T) {
	tc := []struct {
		name    string
		alerts  template.Data
		err     error
		tracked []float64
	}{
		{
			name: "with alerts to track",
			alerts: template.Data{
				Alerts: template.Alerts{
					template.Alert{
						Labels:      template.KV{model.AlertNameLabel: "e2ealertingAlwaysFiring"},
						Annotations: template.KV{"time": "1.604069614e+09"},
						Status:      string(model.AlertFiring),
					},
					template.Alert{
						Labels:      template.KV{model.AlertNameLabel: "e2ealertingAlwaysFiring"},
						Annotations: template.KV{"time": "1.604069615e+09"},
						Status:      string(model.AlertFiring),
					},
				},
			},
			tracked: []float64{1604069614.00, 1604069615.00},
		},
		{
			name: "with alerts that don't have a time annotation or alertname label it ignores them",
			alerts: template.Data{
				Alerts: template.Alerts{
					template.Alert{
						Labels:      template.KV{model.AlertNameLabel: "e2ealertingAlwaysFiring"},
						Annotations: template.KV{"time": "1.604069614e+09"},
						Status:      string(model.AlertFiring),
					},
					template.Alert{
						Labels: template.KV{model.AlertNameLabel: "e2ealertingAlwaysFiring"},
						Status: string(model.AlertFiring),
					},
					template.Alert{
						Annotations: template.KV{"time": "1.604069614e+09"},
						Status:      string(model.AlertFiring),
					},
				},
			},
			tracked: []float64{1604069614.00},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			r, err := NewReceiver(
				ReceiverConfig{
					PurgeInterval:          1 * time.Hour,
					CustomHistogramBuckets: []string{"0.5", "1", "10"},
				},
				log.NewNopLogger(),
				reg,
			)
			require.NoError(t, err)

			router := mux.NewRouter()
			r.RegisterRoutes(router)

			b, err := json.Marshal(tt.alerts)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/receiver", bytes.NewBuffer(b))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, len(tt.tracked), len(r.timestamps))
			for _, timestamp := range tt.tracked {
				_, exists := r.timestamps[timestamp]
				require.True(t, exists, fmt.Sprintf("time %f is not tracked", timestamp))
			}

			metrics, err := reg.Gather()
			require.NoError(t, err)
			for _, metricFamily := range metrics {
				if strings.Contains(*metricFamily.Name, "end_to_end_duration_seconds") {
					for _, metric := range metricFamily.GetMetric() {
						require.Len(t, metric.Histogram.Bucket, 3)
						bucketBounds := []float64{}
						for _, bucket := range metric.Histogram.Bucket {
							bucketBounds = append(bucketBounds, *bucket.UpperBound)
						}
						require.Equal(t, []float64{0.5, 1, 10}, bucketBounds)
					}
				}
			}
		})
	}
}
