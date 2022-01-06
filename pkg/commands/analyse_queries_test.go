package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetInsert(t *testing.T) {
	for _, tc := range []struct {
		initial  []string
		value    string
		expected []string
	}{
		{
			initial:  []string{},
			value:    "foo",
			expected: []string{"foo"},
		},
		{
			initial:  []string{"foo"},
			value:    "foo",
			expected: []string{"foo"},
		},
		{
			initial:  []string{"foo"},
			value:    "bar",
			expected: []string{"bar", "foo"},
		},
		{
			initial:  []string{"bar"},
			value:    "foo",
			expected: []string{"bar", "foo"},
		},
		{
			initial:  []string{"bar", "foo"},
			value:    "bar",
			expected: []string{"bar", "foo"},
		},
	} {
		setInsert(tc.value, &tc.initial)
		require.Equal(t, tc.initial, tc.expected)
	}
}

func TestProcessQuery(t *testing.T) {
	for _, tc := range []struct {
		query    string
		expected map[string]MetricUsage
	}{
		{
			query: `sum(rate(requests_total{status=~"5.."}[5m])) / sum(rate(requests_total[5m]))`,
			expected: map[string]MetricUsage{
				"requests_total": {LabelsUsed: []string{"status"}},
			},
		},
		{
			query: `sum(rate(requests_sum[5m])) / sum(rate(requests_total[5m]))`,
			expected: map[string]MetricUsage{
				"requests_total": {LabelsUsed: nil},
				"requests_sum":   {LabelsUsed: nil},
			},
		},
		{
			query: `sum by (path) (rate(requests_total{status=~"5.."}[5m]))`,
			expected: map[string]MetricUsage{
				"requests_total": {LabelsUsed: []string{"path", "status"}},
			},
		},
	} {
		actual := map[string]MetricUsage{}
		err := processQuery(tc.query, actual)
		require.NoError(t, err)
		require.Equal(t, tc.expected, actual)
	}
}
