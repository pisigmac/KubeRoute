package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Config struct {
	JWTSecret          string
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
}

type Service struct {
	db     *pgxpool.Pool
	config Config
	ghOAuth *oauth2.Config
	googleOAuth *oauth2.Config
}

func NewService(db *pgxpool.Pool, config Config) *Service {
	return &Service{
		db: db,
		config: config,
		ghOAuth: &oauth2.Config{
			ClientID:     config.GitHubClientID,
			ClientSecret: config.GitHubClientSecret,
			Endpoint:     github.Endpoint,
			Scopes:       []string{"read:user", "user:email"},
		},
		googleOAuth: &oauth2.Config{
			ClientID:     config.GoogleClientID,
			ClientSecret: config.GoogleClientSecret,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"openid", "email", "profile"},
		},
	}
}

type User struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Provider    string    `json:"provider"`
	ProviderID  string    `json:"provider_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role"`
}

func (s *Service) GenerateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *Service) GitHubAuthURL(state string) string {
	return s.ghOAuth.AuthCodeURL(state)
}

func (s *Service) GoogleAuthURL(state string) string {
	return s.googleOAuth.AuthCodeURL(state)
}

func (s *Service) HandleGitHubCallback(ctx context.Context, code string) (*User, string, error) {
	token, err := s.ghOAuth.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("github token exchange failed: %w", err)
	}

	// Fetch user info from GitHub
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var ghUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Name      string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return nil, "", err
	}

	// Get email if not public
	email := ghUser.Email
	if email == "" {
		req2, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
		req2.Header.Set("Authorization", "Bearer "+token.AccessToken)
		resp2, _ := http.DefaultClient.Do(req2)
		if resp2 != nil {
			defer resp2.Body.Close()
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			json.NewDecoder(resp2.Body).Decode(&emails)
			for _, e := range emails {
				if e.Primary && e.Verified {
					email = e.Email
					break
				}
			}
		}
	}

	return s.upsertUser(ctx, "github", fmt.Sprintf("%d", ghUser.ID), email, ghUser.Name, ghUser.AvatarURL)
}

func (s *Service) HandleGoogleCallback(ctx context.Context, code string) (*User, string, error) {
	token, err := s.googleOAuth.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("google token exchange failed: %w", err)
	}

	// Fetch user info from Google
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var gUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gUser); err != nil {
		return nil, "", err
	}

	return s.upsertUser(ctx, "google", gUser.ID, gUser.Email, gUser.Name, gUser.Picture)
}

func (s *Service) upsertUser(ctx context.Context, provider, providerID, email, name, avatar string) (*User, string, error) {
	var user User
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, provider, provider_id, email, display_name, avatar_url, role
		FROM users WHERE provider = $1 AND provider_id = $2
	`, provider, providerID).Scan(
		&user.ID, &user.TenantID, &user.Provider, &user.ProviderID,
		&user.Email, &user.DisplayName, &user.AvatarURL, &user.Role,
	)
	if err == nil {
		jwtToken, err := s.GenerateJWT(user)
		return &user, jwtToken, err
	}

	tenantID := uuid.New()
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	if slug == "" {
		slug = "dev"
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO tenants (id, name, slug, plan, billing_email, task_credits)
		VALUES ($1, $2, $3, 'free', $4, 100)
		ON CONFLICT (id) DO NOTHING
	`, tenantID, name+"'s Team", slug+"-"+tenantID.String()[:8], email)
	if err != nil {
		return nil, "", err
	}

	err = s.db.QueryRow(ctx, `
		INSERT INTO users (tenant_id, provider, provider_id, email, display_name, avatar_url, role)
		VALUES ($1, $2, $3, $4, $5, $6, 'owner')
		ON CONFLICT (provider, provider_id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url
		RETURNING id, tenant_id, provider, provider_id, email, display_name, avatar_url, role
	`, tenantID, provider, providerID, email, name, avatar).Scan(
		&user.ID, &user.TenantID, &user.Provider, &user.ProviderID,
		&user.Email, &user.DisplayName, &user.AvatarURL, &user.Role,
	)
	if err != nil {
		return nil, "", err
	}

	jwtToken, err := s.GenerateJWT(user)
	if err != nil {
		return nil, "", err
	}
	return &user, jwtToken, nil
}

func ResolveJWTSecret(env, secret string) (string, error) {
	local := env == "" || env == "local" || env == "development"
	if !local {
		switch secret {
		case "", "dev-secret", "dev-secret-change-in-production":
			return "", fmt.Errorf("JWT_SECRET must be a non-default value when KUBEROUTE_ENV=%s", env)
		}
	}
	if secret == "" {
		return "dev-secret", nil
	}
	return secret, nil
}

func (s *Service) AllowDevAuth(env string) bool {
	return env == "" || env == "local" || env == "development"
}

func (s *Service) DevLogin(ctx context.Context) (*User, string, error) {
	return s.upsertUser(ctx, "github", "dev-local", "dev@localhost", "Local Dev", "")
}

func (s *Service) TenantCredits(ctx context.Context, tenantID uuid.UUID) (plan string, credits, used int, err error) {
	err = s.db.QueryRow(ctx, `SELECT plan, task_credits, task_credits_used FROM tenants WHERE id = $1`, tenantID).
		Scan(&plan, &credits, &used)
	return
}

func ConstantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func (s *Service) GenerateJWT(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       user.ID.String(),
		"tenant_id": user.TenantID.String(),
		"email":     user.Email,
		"role":      user.Role,
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) ValidateJWT(tokenString string) (*User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	userID, _ := uuid.Parse(claims["sub"].(string))
	tenantID, _ := uuid.Parse(claims["tenant_id"].(string))

	return &User{
		ID:       userID,
		TenantID: tenantID,
		Email:    claims["email"].(string),
		Role:     claims["role"].(string),
	}, nil
}

func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error":"invalid auth header"}`, http.StatusUnauthorized)
			return
		}

		user, err := s.ValidateJWT(parts[1])
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type userContextKey struct{}

func UserFromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userContextKey{}).(*User)
	return u
}
