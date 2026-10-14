package executor

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	pb "github.com/kuberoute/agent/proto"
)

type Executor struct {
	client kubernetes.Interface
}

func New(client kubernetes.Interface) *Executor {
	return &Executor{client: client}
}

func (e *Executor) Apply(ctx context.Context, d *pb.RoutingDecision) error {
	if e.client == nil {
		return fmt.Errorf("kubernetes client not configured")
	}
	switch d.DecisionType {
	case "placement":
		ns := d.WorkloadNamespace
		if ns == "" {
			ns = d.Params["namespace"]
		}
		if ns == "" {
			ns = "default"
		}
		kind := d.Params["kind"]
		if kind == "" {
			kind = "Deployment"
		}
		op := d.Params["op"]
		if op == "unpin" {
			return e.unpinWorkload(ctx, ns, d.WorkloadName, kind)
		}
		if d.TargetNode == "" {
			return fmt.Errorf("placement decision %s missing target node", d.DecisionId)
		}
		return e.pinWorkload(ctx, ns, d.WorkloadName, kind, d.TargetNode)
	case "scaling":
		return fmt.Errorf("scaling decisions are not applied in this loop")
	default:
		return fmt.Errorf("unknown decision type: %s", d.DecisionType)
	}
}

func (e *Executor) pinWorkload(ctx context.Context, ns, name, kind, node string) error {
	if name == "" {
		return fmt.Errorf("workload name required")
	}
	affinity := hostnameAffinity(node)
	switch strings.ToLower(kind) {
	case "deployment", "replicaset":
		dep, err := e.client.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get deployment %s/%s: %w", ns, name, err)
		}
		if dep.Spec.Template.Spec.Affinity == nil {
			dep.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		dep.Spec.Template.Spec.Affinity.NodeAffinity = affinity.NodeAffinity
		_, err = e.client.AppsV1().Deployments(ns).Update(ctx, dep, metav1.UpdateOptions{})
		return err
	case "statefulset":
		ss, err := e.client.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get statefulset %s/%s: %w", ns, name, err)
		}
		if ss.Spec.Template.Spec.Affinity == nil {
			ss.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		}
		ss.Spec.Template.Spec.Affinity.NodeAffinity = affinity.NodeAffinity
		_, err = e.client.AppsV1().StatefulSets(ns).Update(ctx, ss, metav1.UpdateOptions{})
		return err
	default:
		return fmt.Errorf("cannot pin kind %s", kind)
	}
}

func (e *Executor) unpinWorkload(ctx context.Context, ns, name, kind string) error {
	if name == "" {
		return fmt.Errorf("workload name required")
	}
	switch strings.ToLower(kind) {
	case "deployment", "replicaset":
		dep, err := e.client.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get deployment %s/%s: %w", ns, name, err)
		}
		if dep.Spec.Template.Spec.Affinity != nil {
			dep.Spec.Template.Spec.Affinity.NodeAffinity = nil
			if affinityEmpty(dep.Spec.Template.Spec.Affinity) {
				dep.Spec.Template.Spec.Affinity = nil
			}
		}
		_, err = e.client.AppsV1().Deployments(ns).Update(ctx, dep, metav1.UpdateOptions{})
		return err
	case "statefulset":
		ss, err := e.client.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("get statefulset %s/%s: %w", ns, name, err)
		}
		if ss.Spec.Template.Spec.Affinity != nil {
			ss.Spec.Template.Spec.Affinity.NodeAffinity = nil
			if affinityEmpty(ss.Spec.Template.Spec.Affinity) {
				ss.Spec.Template.Spec.Affinity = nil
			}
		}
		_, err = e.client.AppsV1().StatefulSets(ns).Update(ctx, ss, metav1.UpdateOptions{})
		return err
	default:
		return fmt.Errorf("cannot unpin kind %s", kind)
	}
}

func affinityEmpty(a *corev1.Affinity) bool {
	if a == nil {
		return true
	}
	return a.NodeAffinity == nil && a.PodAffinity == nil && a.PodAntiAffinity == nil
}

func hostnameAffinity(node string) *corev1.Affinity {
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{{
						Key:      "kubernetes.io/hostname",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{node},
					}},
				}},
			},
		},
	}
}
