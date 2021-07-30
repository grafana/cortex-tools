package commands

import (
	"testing"

	"github.com/grafana/cortex-tools/pkg/analyse"
	"github.com/grafana/cortex-tools/pkg/rules"
	"github.com/stretchr/testify/assert"
)

func TestParseMetricsInRuleFile(t *testing.T) {
	var metrics = []string{
		"apiserver_request_duration_seconds_bucket",
		"apiserver_request_duration_seconds_count",
		"apiserver_request_total",
		"container_cpu_usage_seconds_total",
		"container_memory_cache",
		"container_memory_rss",
		"container_memory_swap",
		"container_memory_working_set_bytes",
		"kube_pod_container_resource_limits",
		"kube_pod_container_resource_requests",
		"kube_pod_info",
		"kube_pod_owner",
		"kube_pod_status_phase",
		"kube_replicaset_owner",
		"kubelet_node_name",
		"kubelet_pleg_relist_duration_seconds_bucket",
		"node_cpu_seconds_total",
		"node_memory_Buffers_bytes",
		"node_memory_Cached_bytes",
		"node_memory_MemAvailable_bytes",
		"node_memory_MemFree_bytes",
		"node_memory_Slab_bytes",
		"scheduler_binding_duration_seconds_bucket",
		"scheduler_e2e_scheduling_duration_seconds_bucket",
		"scheduler_scheduling_algorithm_duration_seconds_bucket",
	}

	output := &analyse.MetricsInRuler{}
	output.OverallMetrics = make(map[string]struct{})

	nss, err := rules.ParseFiles("cortex", []string{"testdata/prometheus_rules.yaml"})
	if err != nil {
		t.Errorf("could not parse rules file")
	}

	for _, ns := range nss {
		for _, group := range ns.Groups {
			err := parseMetricsInRuleGroup(output, group, ns.Namespace)
			if err != nil {
				t.Errorf("could not parse metrics in rule group")
			}
		}
	}
	assert.Equal(t, metrics, output.MetricsUsed)
}
