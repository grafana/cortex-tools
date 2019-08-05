package main

import (
	"os"

	"github.com/grafana/cortex-tool/pkg/commands"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	chunkCommand commands.ChunkCommand
	ruleCommand  commands.RuleCommand
)

func main() {
	log.SetLevel(log.DebugLevel)
	kingpin.Version("0.0.1")
	app := kingpin.New("cortex-cli", "A command-line tool to manage cortex.")
	chunkCommand.Register(app)
	ruleCommand.Register(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
