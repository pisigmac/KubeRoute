package executor

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	pb "github.com/kuberoute/agent/proto"
)

func TestApplyPlacementPatchesDeploymentAffinity(t *testing.T) {
	client := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "default"},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "web"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "web"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Image: "nginx"}}},
			},
		},
	})
	ex := New(client)
	err := ex.Apply(context.Background(), &pb.RoutingDecision{
		DecisionId:        "d1",
		DecisionType:      "placement",
		WorkloadNamespace: "default",
		WorkloadName:      "web",
		TargetNode:        "cheap-1",
		Params:            map[string]string{"kind": "Deployment"},
	})
	if err != nil {
		t.Fatal(err)
	}
	dep, err := client.AppsV1().Deployments("default").Get(context.Background(), "web", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	terms := dep.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	got := terms[0].MatchExpressions[0].Values[0]
	if got != "cheap-1" {
		t.Fatalf("pinned node = %s, want cheap-1", got)
	}
}

func TestApplyUnpinClearsDeploymentAffinity(t *testing.T) {
	client := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "default"},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "web"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "web"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "app", Image: "nginx"}},
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{{
									MatchExpressions: []corev1.NodeSelectorRequirement{{
										Key:      "kubernetes.io/hostname",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"cheap-1"},
									}},
								}},
							},
						},
					},
				},
			},
		},
	})
	ex := New(client)
	err := ex.Apply(context.Background(), &pb.RoutingDecision{
		DecisionId:        "d2",
		DecisionType:      "placement",
		WorkloadNamespace: "default",
		WorkloadName:      "web",
		Params:            map[string]string{"kind": "Deployment", "op": "unpin"},
	})
	if err != nil {
		t.Fatal(err)
	}
	dep, err := client.AppsV1().Deployments("default").Get(context.Background(), "web", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if dep.Spec.Template.Spec.Affinity != nil {
		t.Fatal("expected affinity cleared")
	}
}

func TestApplyRequiresClient(t *testing.T) {
	ex := New(nil)
	err := ex.Apply(context.Background(), &pb.RoutingDecision{DecisionType: "placement", TargetNode: "n1", WorkloadName: "web"})
	if err == nil {
		t.Fatal("expected error")
	}
}
