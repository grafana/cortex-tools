package main

import (
	"os"

	"github.com/grafana/cortex-tool/pkg/commands"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	deleteCommand commands.DeleteChunkCommand
	logConfig     commands.LoggerConfig
)

func main() {
	kingpin.Version("0.0.1")
	app := kingpin.New("chunk-tool", "A command-line tool to manage cortex chunks.")
	logConfig.Register(app)
	deleteCommand.Register(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
