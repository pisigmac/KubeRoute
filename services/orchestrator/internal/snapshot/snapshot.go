package snapshot

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	pb "github.com/kuberoute/orchestrator/proto"
)

type Node struct {
	Name           string  `json:"name"`
	CapacityCPU    float64 `json:"capacity_cpu"`
	CapacityMemory float64 `json:"capacity_memory"`
	AllocatedCPU   float64 `json:"allocated_cpu"`
	AllocatedMemory float64 `json:"allocated_memory"`
	CostPerHour    float64 `json:"cost_per_hour"`
	SpotInstance   bool    `json:"spot_instance"`
	Zone           string  `json:"zone"`
	InstanceType   string  `json:"instance_type"`
}

type Workload struct {
	Name         string  `json:"name"`
	Namespace    string  `json:"namespace"`
	Kind         string  `json:"kind"`
	CPURequest   float64 `json:"cpu_request"`
	MemoryRequest float64 `json:"memory_request"`
	CurrentNode  string  `json:"current_node"`
}

type ClusterSnapshot struct {
	ClusterID uuid.UUID  `json:"cluster_id"`
	Nodes     []Node     `json:"nodes"`
	Workloads []Workload `json:"workloads"`
}

func FromMetrics(clusterID uuid.UUID, metrics []*pb.Metric) ClusterSnapshot {
	nodes := map[string]*Node{}
	workloads := map[string]*Workload{}

	for _, m := range metrics {
		if m == nil {
			continue
		}
		switch m.Name {
		case "node_capacity_cpu", "node_capacity_memory", "node_allocated_cpu", "node_allocated_memory":
			name := m.Labels["node"]
			if name == "" {
				continue
			}
			n := nodes[name]
			if n == nil {
				n = &Node{Name: name}
				nodes[name] = n
			}
			if m.Labels["spot"] == "true" {
				n.SpotInstance = true
			}
			if z := m.Labels["zone"]; z != "" {
				n.Zone = z
			}
			if it := m.Labels["instance_type"]; it != "" {
				n.InstanceType = it
			}
			if c, err := strconv.ParseFloat(m.Labels["cost_per_hour"], 64); err == nil {
				n.CostPerHour = c
			}
			switch m.Name {
			case "node_capacity_cpu":
				n.CapacityCPU = m.Value
			case "node_capacity_memory":
				n.CapacityMemory = m.Value
			case "node_allocated_cpu":
				n.AllocatedCPU = m.Value
			case "node_allocated_memory":
				n.AllocatedMemory = m.Value
			}
		case "workload_cpu_request", "workload_memory_request":
			key := m.Labels["namespace"] + "/" + m.Labels["name"]
			if m.Labels["name"] == "" {
				continue
			}
			w := workloads[key]
			if w == nil {
				kind := m.Labels["kind"]
				if kind == "" {
					kind = "Deployment"
				}
				w = &Workload{
					Name:      m.Labels["name"],
					Namespace: m.Labels["namespace"],
					Kind:      kind,
				}
				workloads[key] = w
			}
			if m.Labels["node"] != "" {
				w.CurrentNode = m.Labels["node"]
			}
			if m.Name == "workload_cpu_request" {
				w.CPURequest += m.Value
			} else {
				w.MemoryRequest += m.Value
			}
		}
	}

	snap := ClusterSnapshot{ClusterID: clusterID}
	for _, n := range nodes {
		if n.CostPerHour == 0 {
			if n.SpotInstance {
				n.CostPerHour = 0.02
			} else {
				n.CostPerHour = 0.05
			}
		}
		snap.Nodes = append(snap.Nodes, *n)
	}
	for _, w := range workloads {
		if w.Namespace == "kube-system" || strings.EqualFold(w.Kind, "DaemonSet") {
			continue
		}
		snap.Workloads = append(snap.Workloads, *w)
	}
	return snap
}

func Persist(ctx context.Context, db *pgxpool.Pool, snap ClusterSnapshot) error {
	for _, n := range snap.Nodes {
		labels, _ := json.Marshal(map[string]string{"zone": n.Zone, "instance_type": n.InstanceType})
		_, err := db.Exec(ctx, `
			INSERT INTO nodes (cluster_id, node_name, instance_type, capacity_cpu, capacity_memory, labels, spot_instance, zone, cost_per_hour)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (cluster_id, node_name) DO UPDATE SET
				instance_type = EXCLUDED.instance_type,
				capacity_cpu = EXCLUDED.capacity_cpu,
				capacity_memory = EXCLUDED.capacity_memory,
				labels = EXCLUDED.labels,
				spot_instance = EXCLUDED.spot_instance,
				zone = EXCLUDED.zone,
				cost_per_hour = EXCLUDED.cost_per_hour
		`, snap.ClusterID, n.Name, n.InstanceType, n.CapacityCPU, n.CapacityMemory, labels, n.SpotInstance, n.Zone, n.CostPerHour)
		if err != nil {
			return err
		}
	}
	for _, w := range snap.Workloads {
		ns := w.Namespace
		if ns == "" {
			ns = "default"
		}
		kind := w.Kind
		if kind == "" || kind == "Pod" || kind == "ReplicaSet" {
			kind = "Deployment"
		}
		reqs, _ := json.Marshal(map[string]float64{"cpu": w.CPURequest, "memory": w.MemoryRequest})
		_, err := db.Exec(ctx, `
			INSERT INTO workloads (cluster_id, namespace, name, kind, resource_requests)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (cluster_id, namespace, name) DO UPDATE SET
				kind = EXCLUDED.kind,
				resource_requests = EXCLUDED.resource_requests
		`, snap.ClusterID, ns, w.Name, kind, reqs)
		if err != nil {
			return err
		}
	}
	return nil
}

func NodeIDByName(ctx context.Context, db *pgxpool.Pool, clusterID uuid.UUID, name string) *uuid.UUID {
	var id uuid.UUID
	err := db.QueryRow(ctx, `SELECT id FROM nodes WHERE cluster_id = $1 AND node_name = $2`, clusterID, name).Scan(&id)
	if err != nil {
		return nil
	}
	return &id
}

func WorkloadIDByName(ctx context.Context, db *pgxpool.Pool, clusterID uuid.UUID, ns, name string) *uuid.UUID {
	if ns == "" {
		ns = "default"
	}
	var id uuid.UUID
	err := db.QueryRow(ctx, `SELECT id FROM workloads WHERE cluster_id = $1 AND namespace = $2 AND name = $3`, clusterID, ns, name).Scan(&id)
	if err != nil {
		return nil
	}
	return &id
}
