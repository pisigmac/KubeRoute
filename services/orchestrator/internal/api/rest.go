package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/kuberoute/orchestrator/internal/auth"
	"github.com/kuberoute/orchestrator/internal/clusters"
	"github.com/kuberoute/orchestrator/internal/decisions"
)

type RESTHandler struct {
	authService     *auth.Service
	clusterService  *clusters.Service
	decisionService *decisions.Service
	dashboardURL    string
	env             string
}

func NewRESTHandler(authService *auth.Service, clusterService *clusters.Service, decisionService *decisions.Service) *RESTHandler {
	dash := os.Getenv("DASHBOARD_URL")
	if dash == "" {
		dash = "http://localhost:3000"
	}
	env := os.Getenv("KUBEROUTE_ENV")
	if env == "" {
		env = "local"
	}
	return &RESTHandler{
		authService:     authService,
		clusterService:  clusterService,
		decisionService: decisionService,
		dashboardURL:    dash,
		env:             env,
	}
}

func (h *RESTHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/health", h.Health).Methods("GET")
	r.HandleFunc("/auth/github", h.GitHubAuth).Methods("GET")
	r.HandleFunc("/auth/github/callback", h.GitHubCallback).Methods("GET")
	r.HandleFunc("/auth/google", h.GoogleAuth).Methods("GET")
	r.HandleFunc("/auth/google/callback", h.GoogleCallback).Methods("GET")
	r.HandleFunc("/auth/dev", h.DevAuth).Methods("POST", "GET")

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(h.authService.Middleware)

	api.HandleFunc("/me", h.Me).Methods("GET")
	api.HandleFunc("/clusters", h.ListClusters).Methods("GET")
	api.HandleFunc("/clusters", h.CreateCluster).Methods("POST")
	api.HandleFunc("/clusters/{id}", h.GetCluster).Methods("GET")
	api.HandleFunc("/clusters/{id}", h.DeleteCluster).Methods("DELETE")
	api.HandleFunc("/clusters/{id}/decisions", h.ListDecisions).Methods("GET")
	api.HandleFunc("/clusters/{id}/decisions/{decision_id}", h.GetDecision).Methods("GET")
	api.HandleFunc("/clusters/{id}/decisions/{decision_id}/apply", h.ApplyDecision).Methods("POST")
	api.HandleFunc("/clusters/{id}/decisions/{decision_id}/reject", h.RejectDecision).Methods("POST")
	api.HandleFunc("/clusters/{id}/decisions/{decision_id}/revert", h.RevertDecision).Methods("POST")
	api.HandleFunc("/clusters/{id}/costs", h.ClusterCosts).Methods("GET")
	api.HandleFunc("/costs", h.Costs).Methods("GET")
}

func (h *RESTHandler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *RESTHandler) GitHubAuth(w http.ResponseWriter, r *http.Request) {
	state := h.authService.GenerateState()
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: state, Path: "/", HttpOnly: true, MaxAge: 600, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, h.authService.GitHubAuthURL(state), http.StatusTemporaryRedirect)
}

func (h *RESTHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if !h.validOAuthState(r) {
		respondError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	code := r.URL.Query().Get("code")
	user, token, err := h.authService.HandleGitHubCallback(r.Context(), code)
	if err != nil {
		respondError(w, http.StatusBadRequest, "auth failed")
		return
	}
	h.finishAuth(w, r, user, token)
}

func (h *RESTHandler) GoogleAuth(w http.ResponseWriter, r *http.Request) {
	state := h.authService.GenerateState()
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: state, Path: "/", HttpOnly: true, MaxAge: 600, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, h.authService.GoogleAuthURL(state), http.StatusTemporaryRedirect)
}

func (h *RESTHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !h.validOAuthState(r) {
		respondError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	code := r.URL.Query().Get("code")
	user, token, err := h.authService.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		respondError(w, http.StatusBadRequest, "auth failed")
		return
	}
	h.finishAuth(w, r, user, token)
}

func (h *RESTHandler) DevAuth(w http.ResponseWriter, r *http.Request) {
	if !h.authService.AllowDevAuth(h.env) {
		respondError(w, http.StatusForbidden, "dev auth disabled")
		return
	}
	user, token, err := h.authService.DevLogin(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "dev auth failed")
		return
	}
	h.finishAuth(w, r, user, token)
}

func (h *RESTHandler) validOAuthState(r *http.Request) bool {
	c, err := r.Cookie("oauth_state")
	if err != nil {
		return false
	}
	return auth.ConstantTimeEqual(c.Value, r.URL.Query().Get("state"))
}

func (h *RESTHandler) finishAuth(w http.ResponseWriter, r *http.Request, user *auth.User, token string) {
	accept := r.Header.Get("Accept")
	wantsJSON := strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html")
	if wantsJSON || r.URL.Query().Get("format") == "json" {
		respondJSON(w, http.StatusOK, map[string]interface{}{"user": user, "token": token})
		return
	}
	http.Redirect(w, r, h.dashboardURL+"/login?token="+url.QueryEscape(token), http.StatusFound)
}

func (h *RESTHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	plan, credits, used, err := h.authService.TenantCredits(r.Context(), user.TenantID)
	if err != nil {
		plan, credits, used = "free", 0, 0
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user": user,
		"tenant": map[string]interface{}{
			"id":                user.TenantID,
			"plan":              plan,
			"task_credits":      credits,
			"task_credits_used": used,
		},
	})
}

func (h *RESTHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	page, limit := parsePage(r)
	list, total, err := h.clusterService.List(r.Context(), user.TenantID, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []clusters.Cluster{}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"clusters": list,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *RESTHandler) CreateCluster(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	var req struct {
		Name     string `json:"name"`
		Provider string `json:"provider"`
		Region   string `json:"region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name required")
		return
	}
	cluster, err := h.clusterService.Register(r.Context(), user.TenantID, req.Name, req.Provider, req.Region)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, cluster)
}

func (h *RESTHandler) GetCluster(w http.ResponseWriter, r *http.Request) {
	c, ok := h.clusterForUser(w, r)
	if !ok {
		return
	}
	respondJSON(w, http.StatusOK, c)
}

func (h *RESTHandler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	c, ok := h.clusterForUser(w, r)
	if !ok {
		return
	}
	if err := h.clusterService.Deregister(r.Context(), c.ID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RESTHandler) ListDecisions(w http.ResponseWriter, r *http.Request) {
	c, ok := h.clusterForUser(w, r)
	if !ok {
		return
	}
	page, limit := parsePage(r)
	status := r.URL.Query().Get("status")
	list, total, err := h.decisionService.List(r.Context(), c.ID, status, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []decisions.Decision{}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"decisions": list,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (h *RESTHandler) GetDecision(w http.ResponseWriter, r *http.Request) {
	c, ok := h.clusterForUser(w, r)
	if !ok {
		return
	}
	decisionID, err := uuid.Parse(mux.Vars(r)["decision_id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid decision id")
		return
	}
	d, err := h.decisionService.Get(r.Context(), decisionID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if d == nil || d.ClusterID != c.ID {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respondJSON(w, http.StatusOK, d)
}

func (h *RESTHandler) ApplyDecision(w http.ResponseWriter, r *http.Request) {
	c, ok := h.clusterForUser(w, r)
	if !ok {
		return
	}
	user := auth.UserFromContext(r.Context())
	decisionID, err := uuid.Parse(mux.Vars(r)["decision_id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid decision id")
		return
	}
	d, err := h.decisionService.Get(r.Context(), decisionID)
	if err != nil || d == nil || d.ClusterID != c.ID {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err := h.decisionService.Apply(r.Context(), decisionID, user.ID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "applied"})
}

func (h *RESTHandler) RejectDecision(w http.ResponseWriter, r *http.Request) {
	c, ok := h.clusterForUser(w, r)
	if !ok {
		return
	}
	user := auth.UserFromContext(r.Context())
	decisionID, err := uuid.Parse(mux.Vars(r)["decision_id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid decision id")
		return
	}
	d, err := h.decisionService.Get(r.Context(), decisionID)
	if err != nil || d == nil || d.ClusterID != c.ID {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err := h.decisionService.Reject(r.Context(), decisionID, user.ID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}

func (h *RESTHandler) RevertDecision(w http.ResponseWriter, r *http.Request) {
	c, ok := h.clusterForUser(w, r)
	if !ok {
		return
	}
	decisionID, err := uuid.Parse(mux.Vars(r)["decision_id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid decision id")
		return
	}
	d, err := h.decisionService.Get(r.Context(), decisionID)
	if err != nil || d == nil || d.ClusterID != c.ID {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	rev, err := h.decisionService.Revert(r.Context(), decisionID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, rev)
}

func (h *RESTHandler) ClusterCosts(w http.ResponseWriter, r *http.Request) {
	c, ok := h.clusterForUser(w, r)
	if !ok {
		return
	}
	total, optimized, savings, breakdown, err := h.decisionService.CostSummary(r.Context(), c.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"cluster_id": c.ID,
		"total":      total,
		"optimized":  optimized,
		"savings":    savings,
		"breakdown":  breakdown,
	})
}

func (h *RESTHandler) Costs(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	list, _, err := h.clusterService.List(r.Context(), user.TenantID, 1, 100)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var total, optimized, savings float64
	breakdown := map[string]float64{}
	for _, c := range list {
		t, o, s, b, err := h.decisionService.CostSummary(r.Context(), c.ID)
		if err != nil {
			continue
		}
		total += t
		optimized += o
		savings += s
		for k, v := range b {
			breakdown[k] += v
		}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total":     total,
		"optimized": optimized,
		"savings":   savings,
		"breakdown": breakdown,
	})
}

func (h *RESTHandler) clusterForUser(w http.ResponseWriter, r *http.Request) (*clusters.Cluster, bool) {
	user := auth.UserFromContext(r.Context())
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid cluster id")
		return nil, false
	}
	c, err := h.clusterService.Get(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return nil, false
	}
	if c == nil || c.TenantID != user.TenantID {
		respondError(w, http.StatusNotFound, "not found")
		return nil, false
	}
	return c, true
}

func parsePage(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	return page, limit
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func respondError(w http.ResponseWriter, code int, msg string) {
	respondJSON(w, code, map[string]string{"error": msg})
}
