package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockAgentServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func newAgentChecker(t *testing.T, addr string) *AgentChecker {
	t.Helper()
	logger := zap.NewNop()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewAgentChecker(client, logger)
	if err != nil {
		t.Fatalf("NewAgentChecker: %v", err)
	}
	return checker
}

func TestNewAgentChecker_NilClient(t *testing.T) {
	_, err := NewAgentChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewAgentChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewAgentChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetAgentInfo_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"version":     "1.15.0",
			"cache_state": "running",
		},
	}
	srv := newMockAgentServer(t, http.StatusOK, payload)
	defer srv.Close()

	checker := newAgentChecker(t, srv.URL)
	info, err := checker.GetAgentInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.15.0" {
		t.Errorf("expected version 1.15.0, got %s", info.Version)
	}
	if info.CacheState != "running" {
		t.Errorf("expected cache_state running, got %s", info.CacheState)
	}
	if info.CheckedAt.IsZero() {
		t.Error("expected CheckedAt to be set")
	}
}

func TestGetAgentInfo_ServerError(t *testing.T) {
	srv := newMockAgentServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	checker := newAgentChecker(t, srv.URL)
	_, err := checker.GetAgentInfo()
	if err == nil {
		t.Fatal("expected error on server error response")
	}
}

func TestGetAgentInfo_NotFound(t *testing.T) {
	srv := newMockAgentServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	checker := newAgentChecker(t, srv.URL)
	_, err := checker.GetAgentInfo()
	if err == nil {
		t.Fatal("expected error on 404 response")
	}
}
