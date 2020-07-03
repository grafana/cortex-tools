package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/grafana/cortex-tools/pkg/commands"
)

var (
	logConfig   commands.LoggerConfig
	pushGateway commands.PushGatewayConfig
)

func main() {
	kingpin.Version("0.0.1")
	app := kingpin.New("cortextool", "A command-line tool to manage cortex chunk backends.")
	logConfig.Register(app)
	commands.RegisterChunkCommands(app)
	pushGateway.Register(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	pushGateway.Stop()
}
