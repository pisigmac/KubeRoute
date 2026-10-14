package collector

import (
	"context"
	"log"
	"strings"
	"sync/atomic"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	pb "github.com/kuberoute/agent/proto"
)

type Collector struct {
	interval  time.Duration
	client    kubernetes.Interface
	nodeCount atomic.Int32
}

func New(interval time.Duration, client kubernetes.Interface) *Collector {
	return &Collector{interval: interval, client: client}
}

func (c *Collector) NodeCount() int32 {
	return c.nodeCount.Load()
}

func (c *Collector) Collect(ctx context.Context, clusterID string) <-chan *pb.MetricBatch {
	out := make(chan *pb.MetricBatch)
	go func() {
		defer close(out)
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			batch, err := c.Scrape(ctx, clusterID)
			if err != nil {
				log.Printf("scrape failed: %v", err)
			} else {
				select {
				case out <- batch:
				case <-ctx.Done():
					return
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return out
}

func (c *Collector) Scrape(ctx context.Context, clusterID string) (*pb.MetricBatch, error) {
	if c.client == nil {
		return &pb.MetricBatch{ClusterId: clusterID, Timestamp: time.Now().Unix()}, nil
	}

	nodes, err := c.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	pods, err := c.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	allocatedCPU := map[string]float64{}
	allocatedMem := map[string]float64{}
	for i := range pods.Items {
		p := &pods.Items[i]
		if p.Status.Phase == corev1.PodSucceeded || p.Status.Phase == corev1.PodFailed {
			continue
		}
		node := p.Spec.NodeName
		for _, ctr := range p.Spec.Containers {
			allocatedCPU[node] += cpuCores(ctr.Resources.Requests)
			allocatedMem[node] += memoryBytes(ctr.Resources.Requests)
		}
	}

	var metrics []*pb.Metric
	seen := map[string]struct{}{}
	for i := range nodes.Items {
		n := &nodes.Items[i]
		seen[n.Name] = struct{}{}
		labels := nodeLabels(n)
		metrics = append(metrics,
			metric("node_capacity_cpu", cpuQuantity(n.Status.Allocatable, corev1.ResourceCPU), labels),
			metric("node_capacity_memory", memoryQuantity(n.Status.Allocatable, corev1.ResourceMemory), labels),
			metric("node_allocated_cpu", allocatedCPU[n.Name], labels),
			metric("node_allocated_memory", allocatedMem[n.Name], labels),
		)
	}

	for i := range pods.Items {
		p := &pods.Items[i]
		if p.Status.Phase == corev1.PodSucceeded || p.Status.Phase == corev1.PodFailed {
			continue
		}
		kind, owner := ownerKindName(p)
		if kind == "" {
			kind, owner = "Pod", p.Name
		}
		var cpuReq, memReq float64
		for _, ctr := range p.Spec.Containers {
			cpuReq += cpuCores(ctr.Resources.Requests)
			memReq += memoryBytes(ctr.Resources.Requests)
		}
		labels := map[string]string{
			"namespace": p.Namespace,
			"name":      owner,
			"kind":      kind,
			"node":      p.Spec.NodeName,
			"pod":       p.Name,
		}
		metrics = append(metrics,
			metric("workload_cpu_request", cpuReq, labels),
			metric("workload_memory_request", memReq, labels),
		)
	}

	c.nodeCount.Store(int32(len(seen)))
	return &pb.MetricBatch{
		ClusterId: clusterID,
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
	}, nil
}

func metric(name string, value float64, labels map[string]string) *pb.Metric {
	return &pb.Metric{Name: name, Value: value, Labels: labels}
}

func nodeLabels(n *corev1.Node) map[string]string {
	spot := "false"
	if n.Labels["node.kubernetes.io/lifecycle"] == "spot" || n.Labels["kuberoute.io/spot"] == "true" {
		spot = "true"
	}
	cost := n.Annotations["kuberoute.io/cost-per-hour"]
	if cost == "" {
		cost = n.Labels["kuberoute.io/cost-per-hour"]
	}
	if cost == "" {
		if spot == "true" {
			cost = "0.02"
		} else {
			cost = "0.05"
		}
	}
	return map[string]string{
		"node":          n.Name,
		"spot":          spot,
		"zone":          n.Labels["topology.kubernetes.io/zone"],
		"instance_type": n.Labels["node.kubernetes.io/instance-type"],
		"cost_per_hour": cost,
	}
}

func ownerKindName(p *corev1.Pod) (string, string) {
	for _, o := range p.OwnerReferences {
		if o.Controller == nil || !*o.Controller {
			continue
		}
		kind := o.Kind
		name := o.Name
		if kind == "ReplicaSet" {
			if i := strings.LastIndex(name, "-"); i > 0 {
				name = name[:i]
			}
			kind = "Deployment"
		}
		return kind, name
	}
	return "", ""
}

func cpuCores(req corev1.ResourceList) float64 {
	q, ok := req[corev1.ResourceCPU]
	if !ok {
		return 0
	}
	return float64(q.MilliValue()) / 1000.0
}

func memoryBytes(req corev1.ResourceList) float64 {
	q, ok := req[corev1.ResourceMemory]
	if !ok {
		return 0
	}
	return float64(q.Value())
}

func cpuQuantity(list corev1.ResourceList, key corev1.ResourceName) float64 {
	q, ok := list[key]
	if !ok {
		return 0
	}
	return float64(q.MilliValue()) / 1000.0
}

func memoryQuantity(list corev1.ResourceList, key corev1.ResourceName) float64 {
	q, ok := list[key]
	if !ok {
		return 0
	}
	return float64(q.Value())
}
