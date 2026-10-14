package load

import (
	"testing"
)

func TestLoad_1000Clusters(t *testing.T) {
	// Simulate 1000 clusters sending metrics
	// Target: < 5s decision latency at p95
	t.Skip("Run with k6: k6 run load_test.js")
}
