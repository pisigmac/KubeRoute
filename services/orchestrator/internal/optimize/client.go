package optimize

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kuberoute/orchestrator/internal/snapshot"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type Decision struct {
	ClusterID            string            `json:"cluster_id"`
	DecisionType         string            `json:"decision_type"`
	Reason               string            `json:"reason"`
	Confidence           float64           `json:"confidence"`
	Params               map[string]string `json:"params"`
	PredictedCostSavings float64           `json:"predicted_cost_savings"`
	SolverTrace          json.RawMessage   `json:"solver_trace"`
	ProposedPatch        json.RawMessage   `json:"proposed_patch"`
}

func (c *Client) Optimize(ctx context.Context, snap snapshot.ClusterSnapshot) ([]Decision, error) {
	body, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/optimize/snapshot", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ai-engine status %d", resp.StatusCode)
	}
	var out struct {
		Decisions []Decision `json:"decisions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Decisions, nil
}
