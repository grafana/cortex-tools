package main

import (
	"os"

	"github.com/grafana/cortex-tool/pkg/commands"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	deleteCommand commands.DeleteChunkCommand
)

func main() {
	log.SetLevel(log.DebugLevel)
	kingpin.Version("0.0.1")
	app := kingpin.New("chunk-tool", "A command-line tool to manage cortex chunks.")
	deleteCommand.Register(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
