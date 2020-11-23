package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateMaxAffectedTenants(t *testing.T) {
	tests := []struct {
		nodeTenants [][]int
		expectedMax int
	}{
		{
			nodeTenants: nil,
			expectedMax: 0,
		},
		{
			nodeTenants: [][]int{
				{1, 2, 3},
				{4, 5, 6},
			},
			expectedMax: 0,
		},
		{
			nodeTenants: [][]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			expectedMax: 0,
		},
		{
			nodeTenants: [][]int{
				{1, 2, 3},
				{1, 5, 6},
				{1, 8, 9},
			},
			expectedMax: 1,
		},
		{
			nodeTenants: [][]int{
				{1, 2, 3},
				{1, 2, 6},
				{1, 2, 9},
			},
			expectedMax: 2,
		},
		{
			nodeTenants: [][]int{
				{1, 2, 3},
				{3, 2, 1},
				{4, 5, 6, 7, 8, 9},
			},
			expectedMax: 3,
		},
		{
			nodeTenants: [][]int{
				{1, 2, 3},
				{1, 2, 3, 4, 5, 6},
				{1, 2, 3, 7, 8, 9},
			},
			expectedMax: 3,
		},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expectedMax, calculateMaxAffectedTenants(tc.nodeTenants))
	}
}
