package main

import "testing"

func TestParseConsoleUpstream(t *testing.T) {
	if _, err := parseConsoleUpstream(""); err == nil {
		t.Fatalf("expected error for empty upstream")
	}
	upstream, err := parseConsoleUpstream("http://localhost:3001/")
	if err != nil {
		t.Fatalf("parseConsoleUpstream() err=%v", err)
	}
	if upstream.Scheme != "http" {
		t.Fatalf("scheme=%q, want http", upstream.Scheme)
	}
	if upstream.Host != "localhost:3001" {
		t.Fatalf("host=%q, want localhost:3001", upstream.Host)
	}
	if _, err := parseConsoleUpstream("localhost:3001"); err == nil {
		t.Fatalf("expected error for missing scheme")
	}
}
