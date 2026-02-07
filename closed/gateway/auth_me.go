package main

import (
	"encoding/json"
	"net/http"

	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
)

func authMeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := auth.IdentityFromContext(r.Context())
		if !ok || identity.Subject == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("{\"error\":\"unauthenticated\"}\n"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id": identity.Subject,
			"email":   identity.Email,
			"roles":   identity.Roles,
		})
	})
}
