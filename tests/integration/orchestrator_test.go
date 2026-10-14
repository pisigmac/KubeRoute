package integration

import (
	"testing"
)

func TestEndToEnd_ClusterRegistration(t *testing.T) {
	// Requires docker-compose up
	t.Skip("Run manually: docker-compose -f docker-compose.dev.yml up && go test ./...")
}

func TestEndToEnd_MetricsStream(t *testing.T) {
	t.Skip("Requires running orchestrator and agent")
}
