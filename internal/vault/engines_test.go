package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockEngineServer(t *testing.T, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if payload != nil {
			json.NewEncoder(w).Encode(payload)
		}
	}))
}

func newEngineChecker(t *testing.T, srv *httptest.Server) *EngineChecker {
	t.Helper()
	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewEngineChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewEngineChecker: %v", err)
	}
	return checker
}

func TestNewEngineChecker_NilClient(t *testing.T) {
	_, err := NewEngineChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewEngineChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://localhost", "tok")
	_, err := NewEngineChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestListEngines_Success(t *testing.T) {
	payload := map[string]interface{}{
		"secret/": map[string]string{"type": "kv", "description": "KV store"},
		"pki/":    map[string]string{"type": "pki", "description": "PKI"},
	}
	srv := newMockEngineServer(t, http.StatusOK, payload)
	defer srv.Close()

	checker := newEngineChecker(t, srv)
	engines, err := checker.ListEngines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(engines) != 2 {
		t.Errorf("expected 2 engines, got %d", len(engines))
	}
}

func TestListEngines_ServerError(t *testing.T) {
	srv := newMockEngineServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	checker := newEngineChecker(t, srv)
	_, err := checker.ListEngines(context.Background())
	if err == nil {
		t.Fatal("expected error on server error")
	}
}
