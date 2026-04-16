package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newMockAuthServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func newAuthChecker(t *testing.T, srv *httptest.Server) *AuthChecker {
	t.Helper()
	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewAuthChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewAuthChecker: %v", err)
	}
	return checker
}

func TestNewAuthChecker_NilClient(t *testing.T) {
	_, err := NewAuthChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewAuthChecker_NilLogger(t *testing.T) {
	srv := newMockAuthServer(t, 200, nil)
	defer srv.Close()
	client, _ := NewClient(srv.URL, "tok")
	_, err := NewAuthChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestLookupSelf_Success(t *testing.T) {
	expire := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"accessor":    "abc123",
			"policies":    []string{"default"},
			"ttl":         7200,
			"renewable":   true,
			"expire_time": expire,
		},
	}
	srv := newMockAuthServer(t, 200, body)
	defer srv.Close()

	checker := newAuthChecker(t, srv)
	status, err := checker.LookupSelf(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", status.Accessor)
	}
	if status.TTL != 2*time.Hour {
		t.Errorf("expected TTL 2h, got %v", status.TTL)
	}
	if !status.Renewable {
		t.Error("expected renewable true")
	}
}

func TestLookupSelf_ServerError(t *testing.T) {
	srv := newMockAuthServer(t, 403, nil)
	defer srv.Close()

	checker := newAuthChecker(t, srv)
	_, err := checker.LookupSelf(context.Background())
	if err == nil {
		t.Fatal("expected error on 403")
	}
}
