package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/grafana-tools/sdk"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/grafana/cortex-tools/pkg/analyse"
)

type GrafanaAnalyseCommand struct {
	address     string
	apiKey      string
	readTimeout time.Duration

	outputFile string
}

func (cmd *GrafanaAnalyseCommand) run(k *kingpin.ParseContext) error {
	var (
		boardLinks []sdk.FoundBoard
		rawBoard   []byte
		board      analyse.Board
		err        error
		output     *analyse.MetricsInGrafana
	)
	output.OverallMetrics = make(map[string]struct{})

	ctx, cancel := context.WithTimeout(context.Background(), cmd.readTimeout)
	defer cancel()

	c, err := sdk.NewClient(cmd.address, cmd.apiKey, sdk.DefaultHTTPClient)
	if err != nil {
		return err
	}

	boardLinks, err = c.SearchDashboards(ctx, "", false)
	if err != nil {
		return err
	}

	for _, link := range boardLinks {
		rawBoard, _, err = c.GetRawDashboardBySlug(ctx, link.URI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s for %s\n", err, link.URI)
			continue
		}

		if err = json.Unmarshal(rawBoard, &board); err != nil {
			fmt.Fprintf(os.Stderr, "%s for %s\n", err, link.URI)
			continue
		}
		analyse.ParseMetricsInBoard(output, board)
	}

	err = writeOut(output, cmd.outputFile)
	if err != nil {
		return err
	}

	return nil
}

func writeOut(mig *analyse.MetricsInGrafana, outputFile string) error {
	var metricsUsed []string
	for metric := range mig.OverallMetrics {
		metricsUsed = append(metricsUsed, metric)
	}
	sort.Strings(metricsUsed)

	mig.MetricsUsed = metricsUsed
	out, err := json.MarshalIndent(mig, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(outputFile, out, os.FileMode(int(0666))); err != nil {
		return err
	}

	return nil
}
