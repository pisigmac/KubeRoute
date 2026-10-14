package auth

import (
	"testing"
)

func TestGenerateJWT(t *testing.T) {
	s := &Service{config: Config{JWTSecret: "test-secret"}}
	user := User{Email: "test@example.com", Role: "owner"}
	token, err := s.GenerateJWT(user)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestValidateJWT(t *testing.T) {
	s := &Service{config: Config{JWTSecret: "test-secret"}}
	user := User{Email: "test@example.com", Role: "owner"}
	token, _ := s.GenerateJWT(user)

	validated, err := s.ValidateJWT(token)
	if err != nil {
		t.Fatal(err)
	}
	if validated.Email != user.Email {
		t.Errorf("expected %s, got %s", user.Email, validated.Email)
	}
}

func TestValidateJWT_Invalid(t *testing.T) {
	s := &Service{config: Config{JWTSecret: "test-secret"}}
	_, err := s.ValidateJWT("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestResolveJWTSecret_LocalDefault(t *testing.T) {
	secret, err := ResolveJWTSecret("local", "")
	if err != nil {
		t.Fatal(err)
	}
	if secret != "dev-secret" {
		t.Fatalf("got %s", secret)
	}
}

func TestResolveJWTSecret_ProductionRejectsDefault(t *testing.T) {
	_, err := ResolveJWTSecret("production", "dev-secret")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConstantTimeEqual(t *testing.T) {
	if !ConstantTimeEqual("abc", "abc") {
		t.Fatal("equal strings")
	}
	if ConstantTimeEqual("abc", "abd") {
		t.Fatal("unequal strings")
	}
}
