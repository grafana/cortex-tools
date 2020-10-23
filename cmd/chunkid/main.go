package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/pkg/labels"

	"github.com/cortexproject/cortex/pkg/chunk"
	"github.com/prometheus/common/model"
)

func main() {
	userID := flag.String("user", "", "user id as string")
	labelSetStr := flag.String("labelSet", "", "full labelset of metric, where name is value of label __name__, labels separated by \":\"")
	fromI := flag.Int64("from", 0, "from timestamp in ms")
	throughI := flag.Int64("through", 0, "through timestamp in ms")
	flag.Parse()

	builder := labels.NewBuilder(nil)
	for _, label := range strings.SplitN(*labelSetStr, ":", 2) {
		lv := strings.SplitN(label, "=", 2)
		if len(lv) != 2 {
			panic(fmt.Sprintf("invalid label/value: %s", lv))
		}

		builder.Set(lv[0], lv[1])
	}

	labels := builder.Labels()
	labelSet := make(model.LabelSet)

	for _, l := range labels {
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
