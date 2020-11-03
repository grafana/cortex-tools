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
	fmt.Println("k, min, max, avg, std dev, %tenants")
	for k := 1; k < 100; k += 1 {
		run(k, log(k))
	}
}

func linear(k int) func(float64) int {
	return func(series float64) int {
		return int(math.Ceil(series / float64(k)))
	}
}

func log(k int) func(float64) int {
	return func(series float64) int {
		return int(math.Ceil(math.Log(series) / float64(k)))
	}
}

func run(k int, sizer func(float64) int) {
	nodeSeries := make([]float64, numReplicas)
	nodeTenants := make([][]int, numReplicas)

	for i := 0; i < numTenants; i++ {
		series := rand.ExpFloat64() * avgSeries
		shards := sizer(series)
		if shards > numReplicas {
			shards = numReplicas
		}

		for j := 0; j < shards; j++ {
			shard := rand.Intn(numReplicas)
			nodeSeries[shard] += series / float64(shards)
			nodeTenants[shard] = append(nodeTenants[shard], i)
		}
	}

	// Count tenants affected by double node outage.
	maxAffectedTenants := 0
	for i := 0; i < numReplicas; i++ {
		for j := 0; j < numReplicas; j++ {
			tenants := 0

			for k := 0; k < len(nodeTenants[i]); k++ {
				for l := 0; l < len(nodeTenants[j]); l++ {
					if nodeTenants[i][k] == nodeTenants[j][l] {
						tenants++
					}
				}
			}

			if tenants > maxAffectedTenants {
				maxAffectedTenants = tenants
			}
		}
	}

	fmt.Printf("%d, %f, %f, %f, %f, %f\n", k, min(nodeSeries), max(nodeSeries),
		stat.Mean(nodeSeries, nil), stat.StdDev(nodeSeries, nil),
		float64(maxAffectedTenants)/float64(numTenants))
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
