package bench

import (
	"math/rand"
	"time"

	"github.com/nouney/randomstring"
)

var (
	randGen *rand.Rand

	namespaces = []string{
		"node",
		"db",
		"http",
		"cortex",
		"container",
	}
)

func init() {
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func randomLabelValue() string {
	length := 3 + randGen.Intn(15-3)

	return randomstring.Generate(length)
}
