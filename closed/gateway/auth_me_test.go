package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
)

func TestAuthMeHandlerReturnsIdentity(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req = req.WithContext(auth.ContextWithIdentity(req.Context(), auth.Identity{
		Subject: "user-1",
		Email:   "user@example.local",
		Roles:   []string{"admin"},
	}))
	rec := httptest.NewRecorder()

	authMeHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["user_id"] != "user-1" {
		t.Fatalf("user_id=%v, want user-1", payload["user_id"])
	}
}

func TestAuthMeHandlerRequiresSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	authMeHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", rec.Code)
	}
}
