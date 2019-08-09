package main

import (
	"os"

	"github.com/grafana/cortex-tool/pkg/commands"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	chunkCommand commands.DeleteChunkCommand
	ruleCommand  commands.RuleCommand
	logConfig    commands.LoggerConfig
)

func main() {
	kingpin.Version("0.0.1")
	app := kingpin.New("cortex-cli", "A command-line tool to manage cortex.")
	chunkCommand.Register(app)
	ruleCommand.Register(app)
	logConfig.Register(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
