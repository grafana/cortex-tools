package main

import (
	"flag"
	"os"

	"github.com/go-kit/log/level"
	"github.com/grafana/cortex-tools/pkg/alerting"

	util_log "github.com/cortexproject/cortex/pkg/util/log"
	"github.com/grafana/dskit/flagext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"
)

func main() {
	var (
		serverConfig   server.Config
		runnerConfig   alerting.RunnerConfig
		receiverConfig alerting.ReceiverConfig
	)

	flagext.RegisterFlags(&serverConfig)
	flagext.RegisterFlags(&receiverConfig)
	flagext.RegisterFlags(&runnerConfig)
	flag.Parse()

	util_log.InitLogger(&serverConfig)
	logger := util_log.Logger
	serverConfig.Log = logging.GoKit(logger)

	server, err := server.New(serverConfig)
	if err != nil {
		level.Error(logger).Log("msg", "unable to initialize the server", "err", err)
		os.Exit(1)
	}
	defer server.Shutdown()

	runner, err := alerting.NewRunner(runnerConfig, logger)
	if err != nil {
		level.Error(logger).Log("msg", "unable initialize the runner", "err", err)
		os.Exit(1)
	}
	defer runner.Stop()

	// Create and register the metrics with the registry
	runner.Add(alerting.NewGaugeCase("now_in_seconds"))
	prometheus.MustRegister(runner)

	receiver, err := alerting.NewReceiver(receiverConfig, logger, prometheus.DefaultRegisterer)
	if err != nil {
		level.Error(logger).Log("msg", "unable initialize the Alertmanager webhook receiver", "err", err)
		os.Exit(1)
	}

	receiver.RegisterRoutes(server.HTTP)

	err = server.Run()
	if err != nil {
		level.Error(logger).Log("msg", "unable to start the server", "err", err)
		os.Exit(1)
	}
}
