package snapshot

import (
	"testing"

	"github.com/google/uuid"

	pb "github.com/kuberoute/orchestrator/proto"
)

func TestFromMetricsBuildsNodesAndWorkloads(t *testing.T) {
	id := uuid.New()
	snap := FromMetrics(id, []*pb.Metric{
		{Name: "node_capacity_cpu", Value: 4, Labels: map[string]string{"node": "n1", "spot": "true", "cost_per_hour": "0.02"}},
		{Name: "node_allocated_cpu", Value: 0.5, Labels: map[string]string{"node": "n1"}},
		{Name: "workload_cpu_request", Value: 0.5, Labels: map[string]string{"namespace": "default", "name": "web", "kind": "Deployment", "node": "n1"}},
		{Name: "workload_cpu_request", Value: 0.1, Labels: map[string]string{"namespace": "kube-system", "name": "coredns", "kind": "Deployment", "node": "n1"}},
	})
	if len(snap.Nodes) != 1 || snap.Nodes[0].Name != "n1" {
		t.Fatalf("nodes = %+v", snap.Nodes)
	}
	if snap.Nodes[0].CostPerHour != 0.02 || !snap.Nodes[0].SpotInstance {
		t.Fatalf("node = %+v", snap.Nodes[0])
	}
	if len(snap.Workloads) != 1 || snap.Workloads[0].Name != "web" {
		t.Fatalf("workloads = %+v", snap.Workloads)
	}
}
