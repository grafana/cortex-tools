package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/gonum/stat"
)

const (
	numTenants  = 1000
	avgSeries   = 10000
	numReplicas = 10
)

func main() {
	fmt.Println("k, min, max, avg, std dev")
	for k := 1000; k < 100000; k += 1000 {
		run(float64(k))
	}
}

func run(k float64) {
	nodeSeries := make([]float64, numReplicas)
	nodeTenants := make([][]int, numReplicas)

	for i := 0; i < numTenants; i++ {
		series := rand.ExpFloat64() * avgSeries
		shards := int(math.Ceil(series / k))
		if shards > numReplicas {
			shards = numReplicas
		}

		for j := 0; j < shards; j++ {
			shard := rand.Intn(numReplicas)
			nodeSeries[shard] += series / float64(shards)
			nodeTenants[shard] = append(nodeTenants[shard], i)
		}
	}

	// TODO cound nodes with more than two tenants in common.

	fmt.Printf("%.0f, %f, %f, %f, %f\n", k, min(nodeSeries), max(nodeSeries), stat.Mean(nodeSeries, nil), stat.StdDev(nodeSeries, nil))
}

func min(fs []float64) float64 {
	result := math.MaxFloat64
	for _, f := range fs {
		if f < result {
			result = f
		}
	}
	return result
}

func max(fs []float64) float64 {
	result := 0.0
	for _, f := range fs {
		if f > result {
			result = f
		}
	}
	return result
}
