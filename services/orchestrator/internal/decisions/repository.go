package decisions

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Decision struct {
	ID                   uuid.UUID         `json:"id"`
	ClusterID            uuid.UUID         `json:"cluster_id"`
	WorkloadID           *uuid.UUID        `json:"workload_id,omitempty"`
	DecisionType         string            `json:"decision_type"`
	TargetNodeID         *uuid.UUID        `json:"target_node_id,omitempty"`
	ConfidenceScore      float64           `json:"confidence_score"`
	PredictedCostSavings float64           `json:"predicted_cost_savings"`
	PredictedLatencyMs   int               `json:"predicted_latency_ms"`
	Reason               string            `json:"reason"`
	Applied              bool              `json:"applied"`
	AppliedAt            *time.Time        `json:"applied_at,omitempty"`
	AppliedBy            *uuid.UUID        `json:"applied_by,omitempty"`
	Reverted             bool              `json:"reverted"`
	Executed             bool              `json:"executed"`
	ExecutedAt           *time.Time        `json:"executed_at,omitempty"`
	TaskCost             int               `json:"task_cost"`
	CreatedAt            time.Time         `json:"created_at"`
	Params               map[string]string `json:"params,omitempty"`
	WorkloadName         string            `json:"workload_name,omitempty"`
	WorkloadNamespace    string            `json:"workload_namespace,omitempty"`
	WorkloadKind         string            `json:"workload_kind,omitempty"`
	TargetNodeName       string            `json:"target_node_name,omitempty"`
	Status               string            `json:"status"`
	SolverTrace          json.RawMessage   `json:"solver_trace,omitempty"`
	ProposedPatch        json.RawMessage   `json:"proposed_patch,omitempty"`
	RevertOf             *uuid.UUID        `json:"revert_of,omitempty"`
	Timeline             []TimelineEvent   `json:"timeline,omitempty"`
}

type TimelineEvent struct {
	At    *time.Time `json:"at,omitempty"`
	Event string     `json:"event"`
	By    string     `json:"by,omitempty"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, d *Decision) error {
	if d.TaskCost == 0 {
		d.TaskCost = 1
	}
	params, _ := json.Marshal(d.Params)
	trace := d.SolverTrace
	if len(trace) == 0 {
		trace = []byte("{}")
	}
	patch := d.ProposedPatch
	if len(patch) == 0 {
		patch = []byte("{}")
	}
	return r.db.QueryRow(ctx, `
		INSERT INTO routing_decisions (
			cluster_id, workload_id, decision_type, target_node_id,
			confidence_score, predicted_cost_savings, predicted_latency_ms, reason, task_cost,
			params, workload_name, workload_namespace, workload_kind, target_node_name,
			solver_trace, proposed_patch, revert_of)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at
	`, d.ClusterID, d.WorkloadID, d.DecisionType, d.TargetNodeID,
		d.ConfidenceScore, d.PredictedCostSavings, d.PredictedLatencyMs, d.Reason, d.TaskCost,
		params, nullIfEmpty(d.WorkloadName), nullIfEmpty(d.WorkloadNamespace),
		nullIfEmpty(d.WorkloadKind), nullIfEmpty(d.TargetNodeName),
		trace, patch, d.RevertOf).
		Scan(&d.ID, &d.CreatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Decision, error) {
	row := r.db.QueryRow(ctx, decisionSelect+` WHERE id = $1`, id)
	d, err := scanDecision(row)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return d, err
}

func (r *Repository) ListByCluster(ctx context.Context, clusterID uuid.UUID, status string, limit, offset int) ([]Decision, int, error) {
	query := decisionSelect + ` WHERE cluster_id = $1`
	args := []interface{}{clusterID}
	n := 1
	switch status {
	case "pending":
		query += ` AND applied = false AND reverted = false`
	case "applied":
		query += ` AND applied = true AND executed = false AND reverted = false`
	case "ready":
		query += ` AND applied = true AND executed = false AND reverted = false`
	case "executed":
		query += ` AND executed = true`
	case "rejected":
		query += ` AND reverted = true`
	}
	n++
	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(n)
	n++
	query += ` OFFSET $` + strconv.Itoa(n)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var decisions []Decision
	for rows.Next() {
		d, err := scanDecision(rows)
		if err != nil {
			return nil, 0, err
		}
		decisions = append(decisions, *d)
	}

	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM routing_decisions WHERE cluster_id = $1`, clusterID).Scan(&total)
	return decisions, total, nil
}

func (r *Repository) HasPendingRevert(ctx context.Context, origID uuid.UUID) (bool, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM routing_decisions
		WHERE revert_of = $1 AND applied = false AND reverted = false
	`, origID).Scan(&n)
	return n > 0, err
}

func (r *Repository) HasPendingPlacement(ctx context.Context, clusterID uuid.UUID, ns, name, target string) (bool, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM routing_decisions
		WHERE cluster_id = $1 AND workload_namespace = $2 AND workload_name = $3
		  AND target_node_name = $4 AND applied = false AND reverted = false
	`, clusterID, ns, name, target).Scan(&n)
	return n > 0, err
}

func (r *Repository) Apply(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE routing_decisions SET applied = true, applied_at = NOW(), applied_by = $2
		WHERE id = $1 AND applied = false AND reverted = false
	`, id, userID)
	return err
}

func (r *Repository) Reject(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE routing_decisions SET reverted = true, reverted_at = NOW(), reverted_by = $2
		WHERE id = $1 AND applied = false
	`, id, userID)
	return err
}

func (r *Repository) MarkExecuted(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE routing_decisions SET executed = true, executed_at = NOW()
		WHERE id = $1 AND applied = true
	`, id)
	return err
}

func (r *Repository) DeductTaskCredits(ctx context.Context, tenantID uuid.UUID, amount int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE tenants SET task_credits = task_credits - $2, task_credits_used = task_credits_used + $2
		WHERE id = $1 AND task_credits >= $2
	`, tenantID, amount)
	return err
}

func (r *Repository) CostSummary(ctx context.Context, clusterID uuid.UUID) (total, optimized, savings float64, breakdown map[string]float64, err error) {
	breakdown = map[string]float64{}
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(cost_per_hour), 0) * 730 FROM nodes WHERE cluster_id = $1
	`, clusterID).Scan(&total)
	if err != nil {
		return
	}
	err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(predicted_cost_savings), 0) FROM routing_decisions
		WHERE cluster_id = $1 AND applied = true
	`, clusterID).Scan(&savings)
	if err != nil {
		return
	}
	optimized = total - savings
	if optimized < 0 {
		optimized = 0
	}
	rows, qerr := r.db.Query(ctx, `
		SELECT COALESCE(decision_type, 'other'), COALESCE(SUM(predicted_cost_savings), 0)
		FROM routing_decisions WHERE cluster_id = $1 AND applied = true
		GROUP BY decision_type
	`, clusterID)
	if qerr != nil {
		err = qerr
		return
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		var v float64
		if err = rows.Scan(&k, &v); err != nil {
			return
		}
		breakdown[k] = v
	}
	return
}

const decisionSelect = `
		SELECT id, cluster_id, workload_id, decision_type, target_node_id,
		       confidence_score, predicted_cost_savings, predicted_latency_ms,
		       reason, applied, applied_at, applied_by, reverted, task_cost, created_at,
		       COALESCE(executed, false), executed_at,
		       COALESCE(params, '{}'::jsonb),
		       COALESCE(workload_name, ''), COALESCE(workload_namespace, ''),
		       COALESCE(workload_kind, ''), COALESCE(target_node_name, ''),
		       COALESCE(solver_trace, '{}'::jsonb), COALESCE(proposed_patch, '{}'::jsonb),
		       revert_of
		FROM routing_decisions`

type scanner interface {
	Scan(dest ...any) error
}

func scanDecision(row scanner) (*Decision, error) {
	var d Decision
	var params []byte
	err := row.Scan(
		&d.ID, &d.ClusterID, &d.WorkloadID, &d.DecisionType, &d.TargetNodeID,
		&d.ConfidenceScore, &d.PredictedCostSavings, &d.PredictedLatencyMs,
		&d.Reason, &d.Applied, &d.AppliedAt, &d.AppliedBy, &d.Reverted, &d.TaskCost, &d.CreatedAt,
		&d.Executed, &d.ExecutedAt, &params,
		&d.WorkloadName, &d.WorkloadNamespace, &d.WorkloadKind, &d.TargetNodeName,
		&d.SolverTrace, &d.ProposedPatch, &d.RevertOf,
	)
	if err != nil {
		return nil, err
	}
	if len(params) > 0 {
		_ = json.Unmarshal(params, &d.Params)
	}
	d.Status = statusOf(d)
	d.Timeline = buildTimeline(d)
	return &d, nil
}

func buildTimeline(d Decision) []TimelineEvent {
	created := d.CreatedAt
	events := []TimelineEvent{{At: &created, Event: "opened"}}
	if d.Applied {
		events = append(events, TimelineEvent{At: d.AppliedAt, Event: "applied"})
	}
	if d.Executed {
		events = append(events, TimelineEvent{At: d.ExecutedAt, Event: "executed", By: "agent"})
	}
	if d.Reverted && !d.Applied {
		events = append(events, TimelineEvent{Event: "rejected"})
	}
	if d.RevertOf != nil {
		events = append(events, TimelineEvent{Event: "revert_of", By: d.RevertOf.String()})
	}
	return events
}

func statusOf(d Decision) string {
	switch {
	case d.Reverted:
		return "rejected"
	case d.Executed:
		return "executed"
	case d.Applied:
		return "applied"
	default:
		return "pending"
	}
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
