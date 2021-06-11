package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/grafana/cortex-tools/pkg/commands"
)

var blockGenCommand commands.BlockGenCommand

func main() {
	kingpin.Version("0.0.1")
	app := kingpin.New("blockgen", "A command-line tool to generate cortex blocks.")
	blockGenCommand.Register(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
