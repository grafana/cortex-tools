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

	ctx := context.Background()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		panic(http.ListenAndServe(":80", nil))
	}()

	err = writeBenchmarker.Run(ctx)
	if err != nil {
		level.Error(logger).Log("msg", "benchmarker failed", "err", err)
		os.Exit(1)
	}
}
