package clusters

import (
	"testing"
)

func TestNormalizeProvider(t *testing.T) {
	cases := map[string]string{
		"kind":     "kind",
		"Kind Cluster": "kind",
		"AWS EKS":  "aws",
		"gke":      "gcp",
		"aks":      "azure",
		"baremetal": "on-prem",
	}
	for in, want := range cases {
		if got := NormalizeProvider(in); got != want {
			t.Errorf("NormalizeProvider(%q)=%q want %q", in, got, want)
		}
	}
}
