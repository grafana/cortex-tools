package commands

import (
	"bufio"
	"io"
	"sort"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

type MetricUsage struct {
	LabelsUsed []string
}

func processQueries(r io.Reader) (map[string]MetricUsage, error) {
	metrics := map[string]MetricUsage{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := processQuery(scanner.Text(), metrics); err != nil {
			return nil, err
		}
	}

	return metrics, scanner.Err()
}

func processQuery(query string, metrics map[string]MetricUsage) error {
	expr, err := parser.ParseExpr(query)
	if err != nil {
		return err
	}

	parser.Inspect(expr, func(node parser.Node, path []parser.Node) error {
		vs, ok := node.(*parser.VectorSelector)
		if !ok {
			return nil
		}

		metricName, ok := getName(vs.LabelMatchers)
		if !ok {
			return nil
		}

		usedLabels := metrics[metricName]

		// Add any label names from the selectors to the list of used labels.
		for _, matcher := range vs.LabelMatchers {
			if matcher.Name == labels.MetricName {
				continue
			}
			setInsert(matcher.Name, &usedLabels.LabelsUsed)
		}

		// Find any aggregations in the path and add grouping labels.
		for _, node := range path {
			ae, ok := node.(*parser.AggregateExpr)
			if !ok {
				continue
			}

			for _, label := range ae.Grouping {
				setInsert(label, &usedLabels.LabelsUsed)
			}
		}
		metrics[metricName] = usedLabels

		return nil
	})

	return nil
}

func getName(matchers []*labels.Matcher) (string, bool) {
	for _, matcher := range matchers {
		if matcher.Name == labels.MetricName && matcher.Type == labels.MatchEqual {
			return matcher.Value, true
		}
	}
	return "", false
}

func setInsert(label string, labels *[]string) {
	i := sort.Search(len(*labels), func(i int) bool { return (*labels)[i] >= label })
	if i < len(*labels) && (*labels)[i] == label {
		// label is present at labels[i]
		return
	}

	// label is not present in labels,
	// but i is the index where it would be inserted.
	*labels = append(*labels, "")
	copy((*labels)[i+1:], (*labels)[i:])
	(*labels)[i] = label
}
