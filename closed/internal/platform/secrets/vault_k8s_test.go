package secrets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeTempJWT(t *testing.T, value string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatalf("write jwt: %v", err)
	}
	return path
}

func TestVaultK8sFetchSuccess(t *testing.T) {
	jwt := "jwt-token"
	authToken := "vault-token"
	secretValue := "super-secret"
	jwtPath := writeTempJWT(t, jwt)

	authCalled := false
	secretCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/kubernetes/login":
			authCalled = true
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if body["role"] != "role" || body["jwt"] != jwt {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"auth": map[string]any{
					"client_token":   authToken,
					"lease_duration": 60,
				},
			})
		case "/v1/secret/data/app":
			secretCalled = true
			if r.Header.Get("X-Vault-Token") != authToken {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"lease_id":       "lease-123",
				"lease_duration": 30,
				"data": map[string]any{
					"data": map[string]any{
						"API_KEY": secretValue,
					},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{
		Provider:      "vault_k8s",
		VaultAddr:     server.URL,
		VaultRole:     "role",
		VaultAuthPath: "auth/kubernetes/login",
		VaultJWTPath:  jwtPath,
		VaultTimeout:  2 * time.Second,
		LeaseTTL:      10 * time.Second,
	}
	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	lease, err := mgr.Fetch(context.Background(), Request{ClassRef: "secret/data/app"})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if !authCalled || !secretCalled {
		t.Fatalf("expected auth and secret calls")
	}
	if lease.Env["API_KEY"] != secretValue {
		t.Fatalf("unexpected secret value")
	}
	if lease.LeaseID != "lease-123" {
		t.Fatalf("unexpected lease id: %s", lease.LeaseID)
	}
	remaining := time.Until(lease.ExpiresAt)
	if remaining < 20*time.Second || remaining > 40*time.Second {
		t.Fatalf("unexpected lease ttl: %s", remaining)
	}
}

func TestVaultK8sFetchUnauthorizedDoesNotLeak(t *testing.T) {
	jwt := "jwt-secret"
	authToken := "vault-secret-token"
	secretLeak := "very-secret-value"
	jwtPath := writeTempJWT(t, jwt)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/kubernetes/login":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"auth": map[string]any{
					"client_token":   authToken,
					"lease_duration": 60,
				},
			})
		case "/v1/secret/data/app":
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(secretLeak))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg := Config{
		Provider:      "vault_k8s",
		VaultAddr:     server.URL,
		VaultRole:     "role",
		VaultAuthPath: "auth/kubernetes/login",
		VaultJWTPath:  jwtPath,
		VaultTimeout:  2 * time.Second,
		LeaseTTL:      10 * time.Second,
	}
	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	_, err = mgr.Fetch(context.Background(), Request{ClassRef: "secret/data/app"})
	if err == nil {
		t.Fatalf("expected error")
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, secretLeak) || strings.Contains(errMsg, authToken) || strings.Contains(errMsg, jwt) {
		t.Fatalf("error leaked secret data: %s", errMsg)
	}
}

func TestVaultK8sAuthTimeout(t *testing.T) {
	jwtPath := writeTempJWT(t, "jwt")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		Provider:      "vault_k8s",
		VaultAddr:     server.URL,
		VaultRole:     "role",
		VaultAuthPath: "auth/kubernetes/login",
		VaultJWTPath:  jwtPath,
		VaultTimeout:  5 * time.Millisecond,
	}
	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	_, err = mgr.Fetch(context.Background(), Request{ClassRef: "secret/data/app"})
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestVaultK8sLeaseExpired(t *testing.T) {
	jwtPath := writeTempJWT(t, "jwt")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/kubernetes/login" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"auth": map[string]any{
				"client_token":   "token",
				"lease_duration": 0,
			},
		})
	}))
	defer server.Close()

	cfg := Config{
		Provider:      "vault_k8s",
		VaultAddr:     server.URL,
		VaultRole:     "role",
		VaultAuthPath: "auth/kubernetes/login",
		VaultJWTPath:  jwtPath,
		VaultTimeout:  2 * time.Second,
		LeaseTTL:      0,
	}
	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	_, err = mgr.Fetch(context.Background(), Request{ClassRef: "secret/data/app"})
	if err == nil {
		t.Fatalf("expected lease expiry error")
	}
}
