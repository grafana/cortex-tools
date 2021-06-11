package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/grafana/cortex-tools/pkg/commands"
)

var (
	logConfig          commands.LoggerConfig
	fakeMetricsCommand commands.FakeMetricsCommand
)

func main() {
	kingpin.Version("0.0.1")
	app := kingpin.New("blockgen", "A command-line tool to generate cortex blocks.")
	logConfig.Register(app)
	fakeMetricsCommand.Register(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
