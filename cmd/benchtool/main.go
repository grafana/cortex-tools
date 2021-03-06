package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"

	"github.com/cortexproject/cortex/pkg/util/flagext"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/grafana/cortex-tools/pkg/bench"
)

var (
	writeBenchConfig bench.WriteBenchConfig
)

func main() {
	flagext.RegisterFlags(&writeBenchConfig)
	flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	writeBenchmarker, err := bench.NewWriteBench(writeBenchConfig, logger, prometheus.DefaultRegisterer)
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

	level.Info(logger).Log("msg", "starting writer-benchmarker")
	err = writeBenchmarker.Run(ctx)
	if err != nil {
		level.Error(logger).Log("msg", "benchmarker failed", "err", err)
		os.Exit(1)
	}
}
