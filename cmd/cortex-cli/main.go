package main

import (
	"os"

	"github.com/grafana/cortex-tool/pkg/commands"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	ruleCommand commands.RuleCommand
	logConfig   commands.LoggerConfig
	pushGateway commands.PushGatewayConfig
)

func main() {
	kingpin.Version("0.0.1")
	app := kingpin.New("cortex-cli", "A command-line tool to manage cortex.")
	commands.RegisterChunkCommands(app)
	ruleCommand.Register(app)
	logConfig.Register(app)
	pushGateway.Register(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	pushGateway.Stop()
}
