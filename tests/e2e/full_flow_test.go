package e2e

import (
	"testing"
)

func TestE2E_FullFlow(t *testing.T) {
	// 1. Register cluster
	// 2. Install agent
	// 3. Stream metrics
	// 4. Generate decision
	// 5. Apply decision
	t.Skip("Full E2E test - run in CI with kind cluster")
}
