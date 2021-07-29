package commands

import (
	"encoding/json"

	"github.com/grafana/cortex-tools/pkg/analyse"
	"gopkg.in/alecthomas/kingpin.v2"
)

type RuleFileAnalyseCommand struct {
	RuleFilesList []string
	outputFile    string
}

func (cmd *RuleFileAnalyseCommand) run(k *kingpin.ParseContext) error {
	output := &analyse.MetricsInRuler{}

	for _, file := range cmd.RuleFilesList {
		var ruleConf analyse.RuleConfig
		buf, err := loadFile(file)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(buf, &ruleConf); err != nil {
			return (err)
		}
		for _, group := range ruleConf.RuleGroups {
			parseMetricsInRuleGroup(output, group, ruleConf.Namespace)
		}
	}

	err := writeOutRuleMetrics(output, cmd.outputFile)
	if err != nil {
		return err
	}

	return nil
}
