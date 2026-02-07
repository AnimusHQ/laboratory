package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
	"github.com/animus-labs/animus-go/closed/internal/platform/requestid"
)

func parseConsoleUpstream(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("console upstream is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid console upstream url: %q", raw)
	}
	return parsed, nil
}

func newConsoleProxy(logger *slog.Logger, upstream *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		origHost := r.Host
		origProto := forwardedProto(r)
		director(r)
		r.Host = upstream.Host
		if r.Header.Get("X-Request-Id") == "" {
			if id, err := requestid.New(); err == nil {
				r.Header.Set("X-Request-Id", id)
			}
		}
		r.Header.Set("X-Forwarded-Host", origHost)
		r.Header.Set("X-Forwarded-Proto", origProto)
		appendForwardedFor(r)
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("console proxy error", "request_id", r.Header.Get("X-Request-Id"), "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("{\"error\":\"console_unavailable\"}\n"))
	}
	proxy.FlushInterval = 100 * time.Millisecond
	return proxy
}

func consoleAuthProxy(logger *slog.Logger, cfg auth.Config, authenticator auth.Authenticator, proxy http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authenticator == nil {
			proxy.ServeHTTP(w, r)
			return
		}
		identity, err := authenticator.Authenticate(r.Context(), r)
		if err != nil {
			if errors.Is(err, auth.ErrUnauthenticated) {
				returnTo := auth.SafeReturnTo(r.URL.RequestURI(), cfg)
				http.Redirect(w, r, "/auth/login?return_to="+url.QueryEscape(returnTo), http.StatusFound)
				return
			}
			logger.Warn("console auth failed", "request_id", r.Header.Get("X-Request-Id"), "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("{\"error\":\"unauthorized\"}\n"))
			return
		}
		ctx := auth.ContextWithIdentity(r.Context(), identity)
		proxy.ServeHTTP(w, r.WithContext(ctx))
	})
}

func forwardedProto(r *http.Request) string {
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		return proto
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func appendForwardedFor(r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	if host == "" {
		return
	}
	prior := r.Header.Get("X-Forwarded-For")
	if prior == "" {
		r.Header.Set("X-Forwarded-For", host)
		return
	}
	r.Header.Set("X-Forwarded-For", prior+", "+host)
}
