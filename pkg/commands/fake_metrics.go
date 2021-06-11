package commands

import (
	"context"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cortexproject/cortex/pkg/storage/bucket"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grafana/cortex-tools/pkg/bench"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/tsdb"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v3"
)

const (
	tsdbChunksDir = "chunks"
	indexFile     = "index"
)

// FakeMetricsCommand is the kingpin command for fake metric generation.
type FakeMetricsCommand struct {
	Series      []bench.SeriesDesc `yaml:"series"`
	Interval    time.Duration      `yaml:"interval"`
	BlockSize   time.Duration      `yaml:"block_size"`
	Concurrency int                `yaml:"concurrency"`
	TmpDir      string             `yaml:"tmp_dir"`
	MinT        int64              `yaml:"min_t"`
	MaxT        int64              `yaml:"max_t"`
	Bucket      bucket.Config      `yaml:"bucket"`
	configFile  string
	logger      log.Logger
}

// Register is used to register the command to a parent command.
func (f *FakeMetricsCommand) Register(app *kingpin.Application) {
	bvCmd := app.Command("fake-metrics", "Generate fake metrics and upload them.").Action(f.fakemetrics)
	bvCmd.Flag("config-file", "configuration file for this tool").Required().StringVar(&f.configFile)
}

func (f *FakeMetricsCommand) fakemetrics(k *kingpin.ParseContext) error {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	content, err := os.ReadFile(f.configFile)
	if err != nil {
		return errors.Wrap(err, "unable to read workload YAML file from the disk")
	}

	err = yaml.Unmarshal(content, &f)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal workload YAML file")
	}

	if f.TmpDir == "" {
		var err error
		f.TmpDir, err = ioutil.TempDir("", "fakemetrics")
		if err != nil {
			return errors.Wrap(err, "failed to create tmp dir")
		}
	}

	level.Info(logger).Log("Using tmp dir: %s", f.TmpDir)

	seriesSet, totalSeriesTypeMap := bench.SeriesDescToSeries(f.Series)
	totalSeries := 0
	for _, typeTotal := range totalSeriesTypeMap {
		totalSeries += typeTotal
	}

	writeWorkLoad := bench.WriteWorkload{
		TotalSeries:        totalSeries,
		TotalSeriesTypeMap: totalSeriesTypeMap,
		Replicas:           1,
		Series:             seriesSet,
	}

	interval := f.Interval.Milliseconds()
	blockSize := f.BlockSize.Milliseconds()

	w, err := tsdb.NewBlockWriter(log.NewNopLogger(), f.TmpDir, blockSize)
	if err != nil {
		return err
	}

	ctx := context.Background()

	level.Info(logger).Log("msg", "Generating data", "minT", f.MinT, "maxT", f.MaxT, "interval", interval)
	currentTs := (int64(f.MinT) + interval - 1) / interval * interval
	currentBlock := currentTs / blockSize
	for ; currentTs <= f.MaxT; currentTs += interval {
		if currentBlock != currentTs/blockSize {
			_, err = w.Flush(ctx)
			if err != nil {
				return err
			}

			w, err = tsdb.NewBlockWriter(log.NewNopLogger(), f.TmpDir, blockSize)
			if err != nil {
				return err
			}
			currentBlock = currentTs / blockSize
		}
		app := w.Appender(ctx)

		timeSeries := writeWorkLoad.GenerateTimeSeries("test1", time.Unix(currentTs/1000, 0))

		for _, s := range timeSeries {
			var ref uint64
			labels := prompbLabelsToLabelsLabels(s.Labels)

			sort.Slice(labels, func(i, j int) bool {
				return strings.Compare(labels[i].Name, labels[j].Name) < 0
			})

			for _, sample := range s.Samples {
				ref, err = app.Append(ref, labels, sample.Timestamp, sample.Value)
				if err != nil {
					return err
				}
			}
		}

		err = app.Commit()
		if err != nil {
			return err
		}
	}

	_, err = w.Flush(ctx)
	return err
}

func prompbLabelsToLabelsLabels(in []prompb.Label) labels.Labels {
	out := make(labels.Labels, len(in))
	for idx := range in {
		out[idx].Name = in[idx].Name
		out[idx].Value = in[idx].Value
	}
	return out
}
