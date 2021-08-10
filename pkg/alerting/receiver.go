package alerting

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/util/flagext"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/model"
)

const (
	namespace      = "e2ealerting"
	alertnameLabel = "alertname"
)

type timestampKey struct {
	timestamp   float64
	alertName   string
	labelValues string
}

// Receiver implements the Alertmanager webhook receiver. It evaluates the received alerts, finds if the alert holds an annonnation with a label of "time", and if it does computes now - time for a total duration.
type Receiver struct {
	logger log.Logger
	cfg    ReceiverConfig

	mtx        sync.Mutex
	timestamps map[timestampKey]struct{}

	quit chan struct{}
	wg   sync.WaitGroup

	registry          prometheus.Registerer
	evalTotal         prometheus.Counter
	failedEvalTotal   prometheus.Counter
	roundtripDuration *prometheus.HistogramVec
}

// ReceiverConfig implements the configuration for the alertmanager webhook receiver
type ReceiverConfig struct {
	LabelsToTrack   flagext.StringSliceCSV
	LabelsToForward flagext.StringSliceCSV
	PurgeInterval   time.Duration
	PurgeLookback   time.Duration
	RoundtripLabel  string
}

// RegisterFlags registers the flags for the alertmanager webhook receiver
func (cfg *ReceiverConfig) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&cfg.RoundtripLabel, "receiver.e2eduration-label", "", "Label name and value in the form 'name=value' to add for the Histogram that observes latency.")
	f.DurationVar(&cfg.PurgeInterval, "receiver.purge-interval", 15*time.Minute, "How often should we purge the in-memory measured timestamps tracker.")
	f.DurationVar(&cfg.PurgeLookback, "receiver.purge-lookback", 2*time.Hour, "Period at which measured timestamps will remain in-memory.")
	f.Var(&cfg.LabelsToTrack, "receiver.labels-to-track", "Additional labels to track seen timestamps by, defaults to labels-to-forward.")
	f.Var(&cfg.LabelsToForward, "receiver.labels-to-forward", "Additional labels to split alerts by.")
}

// NewReceiver returns an alertmanager webhook receiver
func NewReceiver(cfg ReceiverConfig, log log.Logger, reg prometheus.Registerer) (*Receiver, error) {
	constLabels := make(map[string]string, 1)
	if cfg.RoundtripLabel != "" {
		l := strings.Split(cfg.RoundtripLabel, "=")

		if len(l) != 2 {
			return nil, fmt.Errorf("the label is not valid, it must have exactly one name and one value: %s has %d parts", l, len(l))
		}

		constLabels[l[0]] = l[1]
	}

	if len(cfg.LabelsToTrack) == 0 {
		cfg.LabelsToTrack = cfg.LabelsToForward
	}

	labelNames := make([]string, 1+len(cfg.LabelsToForward))
	labelNames[0] = alertnameLabel
	if len(cfg.LabelsToForward) > 0 {
		for i, nm := range cfg.LabelsToForward {
			labelNames[i+1] = nm
		}
	}

	r := &Receiver{
		logger:     log,
		cfg:        cfg,
		timestamps: map[timestampKey]struct{}{},
		registry:   reg,
	}

	r.evalTotal = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "webhook_receiver",
		Name:      "evaluations_total",
		Help:      "The total number of evaluations made by the webhook receiver.",
	})

	r.failedEvalTotal = promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "webhook_receiver",
		Name:      "failed_evaluations_total",
		Help:      "Total number of failed evaluations made by the webhook receiver.",
	})

	r.roundtripDuration = promauto.With(reg).NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   namespace,
		Subsystem:   "webhook_receiver",
		Name:        "end_to_end_duration_seconds",
		Help:        "Time spent (in seconds) from scraping a metric to receiving an alert.",
		Buckets:     []float64{5, 15, 30, 60, 90, 120, 240},
		ConstLabels: constLabels,
	}, labelNames)

	r.wg.Add(1)
	go r.purgeTimestamps()

	return r, nil
}

// RegisterRoutes registers the receiver API routes with the provided router.
func (r *Receiver) RegisterRoutes(router *mux.Router) {
	router.Path("/api/v1/receiver").Methods(http.MethodPost).Handler(http.HandlerFunc(r.measureLatency))
}

func (r *Receiver) measureLatency(w http.ResponseWriter, req *http.Request) {
	data := template.Data{}

	// We want to take a snapshot of time the moment we receive the webhook
	now := time.Now()

	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		level.Error(r.logger).Log("msg", "unable to parse alerts", "err", err)
		r.failedEvalTotal.Inc()
		return
	}

	// We only care about firing alerts as part of this analysis.
	for _, alert := range data.Alerts.Firing() {
		labelsToForward := map[string]string{}
		var name string
		for k, v := range alert.Labels {
			for _, lblName := range r.cfg.LabelsToForward {
				if lblName == k {
					labelsToForward[k] = v
				}
			}

			if k != model.AlertNameLabel {
				continue
			}

			name = v
			labelsToForward[alertnameLabel] = v
		}

		if name == "" {
			level.Debug(r.logger).Log("err", "alerts does not have an alertname label - we can't measure it", "labels", alert.Labels.Names())
			continue
		}

		var t float64
		for k, v := range alert.Annotations {
			if k != "time" {
				continue
			}

			timestamp, err := strconv.ParseFloat(v, 64)
			if err != nil {
				level.Error(r.logger).Log("msg", "failed to parse the timestamp of the alert", "alert", name)
				r.failedEvalTotal.Inc()
				continue
			}

			t = timestamp
		}

		if t == 0.0 {
			level.Debug(r.logger).Log("msg", "alert does not have a `time` annotation - we can't measure it", "labels", alert.Labels.Names(), "annonnations", alert.Annotations.Names())
			continue
		}

		// build a unique key of label values to track
		// to track when a timestamp was already seen
		// preserve order specified to be consistent
		var uniqueLabelKey strings.Builder
		for _, k := range r.cfg.LabelsToTrack {
			for alertKey, v := range alert.Labels {
				if alertKey == k {
					uniqueLabelKey.WriteString(v)
				}
			}
		}

		// fill in any missing Prom labels
		for _, k := range r.cfg.LabelsToForward {
			if _, ok := labelsToForward[k]; !ok {
				labelsToForward[k] = ""
			}
		}

		key := timestampKey{
			timestamp:   t,
			alertName:   name,
			labelValues: uniqueLabelKey.String(),
		}

		latency := now.Unix() - int64(t)
		r.mtx.Lock()
		if _, exists := r.timestamps[key]; exists {
			// We have seen this entry before, skip it.
			level.Debug(r.logger).Log("msg", "entry previously evaluated", "timestamp", t, "alert", name, "labelValues", key.labelValues)
			r.mtx.Unlock()
			continue
		}
		r.timestamps[key] = struct{}{}
		r.mtx.Unlock()

		r.roundtripDuration.With(labelsToForward).Observe(float64(latency))
		level.Info(r.logger).Log("alert", name, "labelValues", key.labelValues, "time", time.Unix(int64(t), 0), "duration_seconds", latency, "status", alert.Status)
		r.evalTotal.Inc()
	}

	w.WriteHeader(http.StatusOK)
}

func (r *Receiver) purgeTimestamps() {
	defer r.wg.Add(-1)

	ticker := time.NewTicker(r.cfg.PurgeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.quit:
			return
		case <-ticker.C:
			deadline := time.Now().Add(-r.cfg.PurgeLookback)
			level.Info(r.logger).Log("msg", "purging timestamps", "deadline", deadline)

			r.mtx.Lock()
			var deleted int
			for k := range r.timestamps {
				// purge entry for the timestamp, when the deadline is after the timestamp t
				if deadline.After(time.Unix(int64(k.timestamp), 0)) {
					delete(r.timestamps, k)
					deleted++
				}
			}
			r.mtx.Unlock()
			level.Info(r.logger).Log("msg", "purging done", "count", deleted)
		}
	}
}
