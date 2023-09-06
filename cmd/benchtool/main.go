package main

import (
	"context"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	logutil "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/flagext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/weaveworks/common/logging"

	"github.com/grafana/cortex-tools/pkg/bench"
)

var (
	benchConfig     bench.Config
	LogLevelConfig  logging.Level
	LogFormatConfig logging.Format
)

func main() {
	flagext.RegisterFlags(&benchConfig, &LogLevelConfig, &LogFormatConfig)
	flag.Parse()

	logger, err := logutil.NewPrometheusLogger(LogLevelConfig, LogFormatConfig)
	if err != nil {
		level.Error(logger).Log("msg", "error initializing logger", "err", err)
		os.Exit(1)
	}

	benchmarkRunner, err := bench.NewBenchRunner(benchConfig, logger, prometheus.DefaultRegisterer)
	if err != nil {
		level.Error(logger).Log("msg", "error initializing benchmarker", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		panic(http.ListenAndServe(":80", nil))
	}()

	level.Info(logger).Log("msg", "starting benchmarker")
	err = benchmarkRunner.Run(ctx)
	if err != nil {
		level.Error(logger).Log("msg", "benchmarker failed", "err", err)
		os.Exit(1)
	}
}
