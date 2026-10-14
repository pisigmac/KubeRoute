package decisions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Service struct {
	repo *Repository
	js   nats.JetStreamContext
}

func NewService(repo *Repository, js nats.JetStreamContext) *Service {
	return &Service{repo: repo, js: js}
}

func (s *Service) Create(ctx context.Context, d *Decision) error {
	if err := s.repo.Create(ctx, d); err != nil {
		return fmt.Errorf("failed to create decision: %w", err)
	}
	if s.js != nil {
		data, _ := json.Marshal(d)
		s.js.Publish("decisions.new", data)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Decision, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, clusterID uuid.UUID, status string, page, limit int) ([]Decision, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return s.repo.ListByCluster(ctx, clusterID, status, limit, offset)
}

func (s *Service) HasPendingPlacement(ctx context.Context, clusterID uuid.UUID, ns, name, target string) (bool, error) {
	return s.repo.HasPendingPlacement(ctx, clusterID, ns, name, target)
}

func (s *Service) Apply(ctx context.Context, id, userID uuid.UUID) error {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if d == nil {
		return fmt.Errorf("decision not found")
	}
	if d.Applied {
		return fmt.Errorf("decision already applied")
	}
	if d.Reverted {
		return fmt.Errorf("decision already rejected")
	}
	if err := s.repo.Apply(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to apply decision: %w", err)
	}
	if s.js != nil {
		s.js.Publish("decisions.applied", []byte(id.String()))
	}
	return nil
}

func (s *Service) Reject(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Reject(ctx, id, userID)
}

func (s *Service) Revert(ctx context.Context, id uuid.UUID) (*Decision, error) {
	orig, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if orig == nil {
		return nil, fmt.Errorf("decision not found")
	}
	if orig.DecisionType != "placement" {
		return nil, fmt.Errorf("only placement decisions can be reverted")
	}
	if orig.Reverted && !orig.Applied {
		return nil, fmt.Errorf("rejected decision cannot be reverted")
	}
	if !orig.Applied && !orig.Executed {
		return nil, fmt.Errorf("apply the decision before reverting")
	}
	pending, err := s.repo.HasPendingRevert(ctx, orig.ID)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, fmt.Errorf("a revert is already pending")
	}

	kind := orig.WorkloadKind
	if kind == "" {
		kind = orig.Params["kind"]
	}
	if kind == "" {
		kind = "Deployment"
	}
	ns := orig.WorkloadNamespace
	if ns == "" {
		ns = orig.Params["namespace"]
	}
	name := orig.WorkloadName
	if name == "" {
		name = orig.Params["workload"]
	}
	backTo := orig.Params["current_node"]
	from := orig.TargetNodeName

	rev := &Decision{
		ClusterID:         orig.ClusterID,
		DecisionType:      "placement",
		WorkloadName:      name,
		WorkloadNamespace: ns,
		WorkloadKind:      kind,
		WorkloadID:        orig.WorkloadID,
		TaskCost:          1,
		RevertOf:          &orig.ID,
		Params: map[string]string{
			"workload":     name,
			"namespace":    ns,
			"kind":         kind,
			"current_node": from,
		},
	}

	if backTo != "" {
		rev.TargetNodeName = backTo
		rev.Params["target_node"] = backTo
		rev.Params["op"] = "pin"
		rev.Reason = fmt.Sprintf("Revert: move %s/%s from %s back to %s", ns, name, from, backTo)
		rev.ProposedPatch = placementPatchJSON(kind, ns, name, backTo, "pin")
	} else {
		rev.Params["op"] = "unpin"
		rev.Reason = fmt.Sprintf("Revert: unpin %s/%s (was pinned to %s)", ns, name, from)
		rev.ProposedPatch = placementPatchJSON(kind, ns, name, "", "unpin")
	}
	rev.SolverTrace = revertTraceJSON(orig, backTo)
	if orig.PredictedCostSavings != 0 {
		rev.PredictedCostSavings = -orig.PredictedCostSavings
	}

	exists, err := s.repo.HasPendingPlacement(ctx, orig.ClusterID, ns, name, rev.TargetNodeName)
	if err != nil {
		return nil, err
	}
	if exists && rev.TargetNodeName != "" {
		return nil, fmt.Errorf("a pending placement already exists for this workload")
	}

	if err := s.Create(ctx, rev); err != nil {
		return nil, err
	}
	return rev, nil
}

func placementPatchJSON(kind, ns, name, node, op string) json.RawMessage {
	type expr struct {
		Key      string   `json:"key"`
		Operator string   `json:"operator"`
		Values   []string `json:"values"`
	}
	type term struct {
		MatchExpressions []expr `json:"matchExpressions"`
	}
	type req struct {
		NodeSelectorTerms []term `json:"nodeSelectorTerms"`
	}
	patch := map[string]any{
		"op":        op,
		"kind":      kind,
		"namespace": ns,
		"name":      name,
	}
	if op == "unpin" {
		patch["nodeAffinity"] = nil
	} else {
		patch["nodeAffinity"] = map[string]any{
			"requiredDuringSchedulingIgnoredDuringExecution": req{
				NodeSelectorTerms: []term{{
					MatchExpressions: []expr{{
						Key:      "kubernetes.io/hostname",
						Operator: "In",
						Values:   []string{node},
					}},
				}},
			},
		}
	}
	b, _ := json.Marshal(patch)
	return b
}

func revertTraceJSON(orig *Decision, chosen string) json.RawMessage {
	body, _ := json.Marshal(map[string]any{
		"objective":   "revert",
		"chosen":      chosen,
		"current":     orig.TargetNodeName,
		"candidates":  []any{},
		"constraints": []string{"revert of " + orig.ID.String()},
		"checks":      []string{"undo previous placement pin"},
		"blast_radius": map[string]any{
			"kind":      orig.WorkloadKind,
			"namespace": orig.WorkloadNamespace,
			"name":      orig.WorkloadName,
			"workloads": 1,
		},
	})
	return body
}

func (s *Service) MarkExecuted(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkExecuted(ctx, id)
}

func (s *Service) CostSummary(ctx context.Context, clusterID uuid.UUID) (float64, float64, float64, map[string]float64, error) {
	return s.repo.CostSummary(ctx, clusterID)
}
