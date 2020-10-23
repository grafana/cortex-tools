package main

import (
	"flag"
	"fmt"

	"github.com/cortexproject/cortex/pkg/chunk"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
)

func main() {
	userID := flag.String("user", "", "user id as string")
	metricName := flag.String("metricName", "", "full metricName")
	fromI := flag.Int64("from", 0, "from timestamp in ms")
	throughI := flag.Int64("through", 0, "through timestamp in ms")
	flag.Parse()

	labels, err := parser.ParseMetric(*metricName)
	if err != nil {
		panic(err)
	}

	labelSet := make(model.LabelSet)
	for _, l := range labels {
		fmt.Println(fmt.Sprintf("setting name \"%s\" with value \"%s\"", l.Name, l.Value))
		labelSet[model.LabelName(l.Name)] = model.LabelValue(l.Value)
	}

	c := chunk.Chunk{
		UserID:      *userID,
		Metric:      labels,
		Fingerprint: labelSet.Fingerprint(),
		From:        model.Time(*fromI),
		Through:     model.Time(*throughI),

		// necessary to get the same chunk id format that we're looking for.
		ChecksumSet: true,
	}

	fmt.Println(fmt.Sprintf("Chunk ID is: %s", c.ExternalKey()))
}
