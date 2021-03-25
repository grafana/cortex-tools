package bench

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/prompb"
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

type QueryDesc struct {
	NumQueries         int           `yaml:"num_queries"`
	ExprTemplate       string        `yaml:"expr_template"`
	RequiredSeriesType SeriesType    `yaml:"series_type"`
	ChurnChance        float64       `yaml:"churn_chance"`
	Interval           time.Duration `yaml:"interval"`
	TimeRange          time.Duration `yaml:"time_range,omitempty"`
}

type WorkloadDesc struct {
	Replicas  int          `yaml:"replicas"`
	Series    []SeriesDesc `yaml:"series"`
	QueryDesc []QueryDesc  `yaml:"queries"`
}

type timeseries struct {
	labelSets  [][]prompb.Label
	lastValue  float64
	seriesType SeriesType
}

type workload struct {
	replicas           int
	series             []*timeseries
	totalSeries        int
	totalSeriesTypeMap map[SeriesType]int
}

func newWorkload(workloadDesc WorkloadDesc, reg prometheus.Registerer) *workload {
	totalSeries := 0
	totalSeriesTypeMap := map[SeriesType]int{
		GaugeZero:     0,
		GaugeRandom:   0,
		CounterOne:    0,
		CounterRandom: 0,
	}

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
			labelSets = addLabelToLabelSet(labelSets, lbl)
		}

		series = append(series, &timeseries{
			labelSets:  labelSets,
			seriesType: seriesDesc.Type,
		})
		numSeries := len(labelSets)
		totalSeries += numSeries
		totalSeriesTypeMap[seriesDesc.Type] += numSeries
	}

	return &workload{
		replicas:           workloadDesc.Replicas,
		series:             series,
		totalSeries:        totalSeries,
		totalSeriesTypeMap: totalSeriesTypeMap,
	}
}

func addLabelToLabelSet(labelSets [][]prompb.Label, lbl LabelDesc) [][]prompb.Label {
	newLabelSets := make([][]prompb.Label, 0, len(labelSets)*lbl.UniqueValues)
	for i := 0; i < lbl.UniqueValues; i++ {
		for _, labelSet := range labelSets {
			newSet := make([]prompb.Label, len(labelSet)+1)
			for i := range labelSet {
				newSet[i] = labelSet[i]
			}
			newSet[len(newSet)-1] = prompb.Label{
				Name:  lbl.Name,
				Value: fmt.Sprintf("%s-%v", lbl.ValuePrefix, i),
			}
			newLabelSets = append(newLabelSets, newSet)
		}
	}
	return newLabelSets
}

func (w *workload) generateTimeSeries(id string, t time.Time) []prompb.TimeSeries {
	now := t.UnixNano() / int64(time.Millisecond)

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
				newLabelSet := make([]prompb.Label, len(labelSet)+1)
				for i := range labelSet {
					newLabelSet[i] = labelSet[i]
				}
				newLabelSet[len(newLabelSet)-1] = replicaLabel
				timeseries = append(timeseries, prompb.TimeSeries{
					Labels: newLabelSet,
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
