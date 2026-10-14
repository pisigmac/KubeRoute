package collector

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestScrapeUsesClusterStateNotRandom(t *testing.T) {
	ctrl := true
	client := fake.NewSimpleClientset(
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "cheap-1",
				Labels: map[string]string{
					"kuberoute.io/spot":          "true",
					"kuberoute.io/cost-per-hour": "0.02",
				},
			},
			Status: corev1.NodeStatus{
				Allocatable: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("4"),
					corev1.ResourceMemory: resource.MustParse("8Gi"),
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "web-abc",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{{
					Kind:       "ReplicaSet",
					Name:       "web-6f8c",
					Controller: &ctrl,
				}},
			},
			Spec: corev1.PodSpec{
				NodeName: "cheap-1",
				Containers: []corev1.Container{{
					Name: "app",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				}},
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
	)

	c := New(time.Hour, client)
	batch, err := c.Scrape(context.Background(), "cluster-1")
	if err != nil {
		t.Fatal(err)
	}
	if c.NodeCount() != 1 {
		t.Fatalf("node count = %d, want 1", c.NodeCount())
	}

	names := map[string]int{}
	for _, m := range batch.Metrics {
		names[m.Name]++
		if m.Name == "node_capacity_cpu" && m.Value != 4 {
			t.Errorf("capacity cpu = %v, want 4", m.Value)
		}
		if m.Name == "node_allocated_cpu" && m.Value != 0.5 {
			t.Errorf("allocated cpu = %v, want 0.5", m.Value)
		}
		if m.Name == "workload_cpu_request" {
			if m.Labels["name"] != "web" {
				t.Errorf("workload name = %s, want web", m.Labels["name"])
			}
			if m.Labels["kind"] != "Deployment" {
				t.Errorf("kind = %s, want Deployment", m.Labels["kind"])
			}
		}
	}
	if names["node_capacity_cpu"] == 0 || names["workload_cpu_request"] == 0 {
		t.Fatalf("missing expected metrics: %v", names)
	}
}

func TestScrapeNilClientEmpty(t *testing.T) {
	c := New(time.Hour, nil)
	batch, err := c.Scrape(context.Background(), "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Metrics) != 0 {
		t.Fatalf("expected no metrics, got %d", len(batch.Metrics))
	}
}
