package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/gonum/stat"
)

const (
	numTenants        = 1000
	avgSeries         = 100000 // After replication
	numReplicas       = 100
	replicationFactor = 3
	function          = "linear"
)

func main() {
	// Print CSV header.
	fmt.Printf(
		"k, min, max, avg, std dev, %% tenants affected by double node outage, setup = %s function / %d tenants / %d avg series per tenant / %d replicas / %dx replication factor\n",
		function, numTenants, avgSeries, numReplicas, replicationFactor)

	switch function {
	case "linear":
		for k := 1000; k <= 100000; k += 1000 {
			run(k, linear(k))
		}
	case "log":
		for k := 1; k <= 100; k++ {
			run(k, log(k))
		}
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

	// Simulate the distribution of tenants across replicas.
	for tenantID := 0; tenantID < numTenants; tenantID++ {
		// Seed the pseudo-random generator with the tenant ID, so that different runs
		// get the same number of series for the same tenant.
		entropy := rand.New(rand.NewSource(int64(tenantID)))

		numSeries := entropy.ExpFloat64() * avgSeries

		shardSize := sizer(numSeries)
		if shardSize > numReplicas {
			shardSize = numReplicas
		} else if shardSize < replicationFactor {
			shardSize = replicationFactor
		}

		replicaIDs := shuffleShard(entropy, shardSize, numReplicas)
		for _, replicaID := range replicaIDs {
			nodeSeries[replicaID] += numSeries / float64(shardSize)
			nodeTenants[replicaID] = append(nodeTenants[replicaID], tenantID)
		}
	}

	// Count tenants affected by double node outage.
	maxAffectedTenants := calculateMaxAffectedTenants(nodeTenants)

	fmt.Printf("%d, %d, %d, %d, %f, %f\n",
		k,
		int(min(nodeSeries)),
		int(max(nodeSeries)),
		int(stat.Mean(nodeSeries, nil)),
		stat.StdDev(nodeSeries, nil),
		float64(maxAffectedTenants)/float64(numTenants))
}

func calculateMaxAffectedTenants(nodeTenants [][]int) int {
	maxAffectedTenants := 0

	for i := 0; i < len(nodeTenants); i++ {
		for j := 0; j < len(nodeTenants); j++ {
			// Skip the same node.
			if i == j {
				continue
			}

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

	return maxAffectedTenants
}

func shuffleShard(entropy *rand.Rand, shardSize, numReplicas int) []int {
	// Randomly pick shardSize different replicas.
	replicas := map[int]struct{}{}
	for len(replicas) < shardSize {
		replicas[entropy.Intn(numReplicas)] = struct{}{}
	}

	// Build the list of replica IDs for this tenant.
	ids := make([]int, 0, len(replicas))
	for id := range replicas {
		ids = append(ids, id)
	}

	return ids
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
