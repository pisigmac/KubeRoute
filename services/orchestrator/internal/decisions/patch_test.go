package decisions

import (
	"encoding/json"
	"testing"
)

func TestPlacementPatchJSONPin(t *testing.T) {
	raw := placementPatchJSON("Deployment", "default", "web", "n2", "pin")
	var patch map[string]any
	if err := json.Unmarshal(raw, &patch); err != nil {
		t.Fatal(err)
	}
	if patch["op"] != "pin" {
		t.Fatalf("op=%v", patch["op"])
	}
	aff := patch["nodeAffinity"].(map[string]any)
	req := aff["requiredDuringSchedulingIgnoredDuringExecution"].(map[string]any)
	terms := req["nodeSelectorTerms"].([]any)
	term := terms[0].(map[string]any)
	exprs := term["matchExpressions"].([]any)
	expr := exprs[0].(map[string]any)
	vals := expr["values"].([]any)
	if vals[0] != "n2" {
		t.Fatalf("values=%v", vals)
	}
}

func TestPlacementPatchJSONUnpin(t *testing.T) {
	raw := placementPatchJSON("Deployment", "default", "web", "", "unpin")
	var patch map[string]any
	if err := json.Unmarshal(raw, &patch); err != nil {
		t.Fatal(err)
	}
	if patch["op"] != "unpin" {
		t.Fatalf("op=%v", patch["op"])
	}
	if patch["nodeAffinity"] != nil {
		t.Fatalf("expected null affinity, got %v", patch["nodeAffinity"])
	}
}
