package commands

import (
	"encoding/json"
	"testing"

	"github.com/grafana-tools/sdk"
	"github.com/stretchr/testify/assert"

	"github.com/grafana/cortex-tools/pkg/analyse"
)

func TestParseMetricsInBoard(t *testing.T) {
	var metrics = []string{
		"apiserver_request:availability30d",
		"apiserver_request_total",
		"cluster_quantile:apiserver_request_duration_seconds:histogram_quantile",
		"code_resource:apiserver_request_total:rate5m",
		"go_goroutines",
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
		"workqueue_adds_total",
		"workqueue_depth",
		"workqueue_queue_duration_seconds_bucket",
	}

	var board sdk.Board
	output := &analyse.MetricsInGrafana{}
	output.OverallMetrics = make(map[string]struct{})

	buf, err := loadFile("testdata/apiserver.json")
	if err != nil {
		t.Errorf("Could not load test dashboard file")
	}
	if err = json.Unmarshal(buf, &board); err != nil {
		t.Errorf("Could not deserialize test dashboard file")
	}

	parseMetricsInBoard(output, board)
	assert.Equal(t, metrics, output.Dashboards[0].Metrics)
}
