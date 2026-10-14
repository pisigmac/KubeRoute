package clusters

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, tenantID uuid.UUID, name, provider, region string) (*Cluster, error) {
	c := &Cluster{
		TenantID: tenantID,
		Name:     name,
		Provider: NormalizeProvider(provider),
		Region:   region,
		Status:   "pending",
		Metadata: make(map[string]interface{}),
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("failed to register cluster: %w", err)
	}
	return c, nil
}

func NormalizeProvider(p string) string {
	p = strings.ToLower(p)
	switch {
	case strings.Contains(p, "aws"), strings.Contains(p, "eks"):
		return "aws"
	case strings.Contains(p, "gcp"), strings.Contains(p, "gke"):
		return "gcp"
	case strings.Contains(p, "azure"), strings.Contains(p, "aks"):
		return "azure"
	case strings.Contains(p, "kind"):
		return "kind"
	default:
		return "on-prem"
	}
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Cluster, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID, page, limit int) ([]Cluster, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.ListByTenant(ctx, tenantID, limit, offset)
}

func (s *Service) Heartbeat(ctx context.Context, id uuid.UUID, nodeCount int, status string) error {
	return s.repo.UpdateHeartbeat(ctx, id, nodeCount, status)
}

func (s *Service) Deregister(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
