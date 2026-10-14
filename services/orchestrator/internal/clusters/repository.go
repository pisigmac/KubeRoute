package clusters

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Cluster struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Name         string    `json:"name"`
	Provider     string    `json:"provider"`
	Region       string    `json:"region"`
	K8sVersion   string    `json:"k8s_version"`
	AgentVersion string    `json:"agent_version"`
	Status       string    `json:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat"`
	NodeCount    int       `json:"node_count"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c *Cluster) error {
	meta, _ := json.Marshal(c.Metadata)
	if c.Metadata == nil {
		meta = []byte("{}")
	}
	return r.db.QueryRow(ctx, `
		INSERT INTO clusters (tenant_id, name, provider, region, k8s_version, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, c.TenantID, c.Name, c.Provider, c.Region, c.K8sVersion, c.Status, meta).
		Scan(&c.ID, &c.CreatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Cluster, error) {
	var c Cluster
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, name, provider, region, k8s_version, agent_version,
		       status, last_heartbeat, node_count, metadata, created_at
		FROM clusters WHERE id = $1
	`, id).Scan(
		&c.ID, &c.TenantID, &c.Name, &c.Provider, &c.Region, &c.K8sVersion,
		&c.AgentVersion, &c.Status, &c.LastHeartbeat, &c.NodeCount, &c.Metadata, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *Repository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]Cluster, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, name, provider, region, k8s_version, agent_version,
		       status, last_heartbeat, node_count, metadata, created_at
		FROM clusters WHERE tenant_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var clusters []Cluster
	for rows.Next() {
		var c Cluster
		rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.Provider, &c.Region, &c.K8sVersion,
			&c.AgentVersion, &c.Status, &c.LastHeartbeat, &c.NodeCount, &c.Metadata, &c.CreatedAt,
		)
		clusters = append(clusters, c)
	}

	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM clusters WHERE tenant_id = $1`, tenantID).Scan(&total)
	return clusters, total, nil
}

func (r *Repository) UpdateHeartbeat(ctx context.Context, id uuid.UUID, nodeCount int, status string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE clusters SET last_heartbeat = NOW(), node_count = $2, status = $3
		WHERE id = $1
	`, id, nodeCount, status)
	return err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM clusters WHERE id = $1`, id)
	return err
}
