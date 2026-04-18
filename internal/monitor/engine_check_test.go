package monitor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockEngineCheckServer(t *testing.T, mounts map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mounts)
	}))
}

func newEngineMonitor(t *testing.T, srv *httptest.Server, expected []string) *EngineMonitor {
	t.Helper()
	client, _ := vault.NewClient(srv.URL, "test-token")
	checker, _ := vault.NewEngineChecker(client, zap.NewNop())
	mon, err := NewEngineMonitor(checker, expected, zap.NewNop())
	if err != nil {
		t.Fatalf("NewEngineMonitor: %v", err)
	}
	return mon
}

func TestNewEngineMonitor_NilChecker(t *testing.T) {
	_, err := NewEngineMonitor(nil, nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewEngineMonitor_NilLogger(t *testing.T) {
	client, _ := vault.NewClient("http://localhost", "tok")
	checker, _ := vault.NewEngineChecker(client, zap.NewNop())
	_, err := NewEngineMonitor(checker, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEngineMonitor_AllPresent(t *testing.T) {
	mounts := map[string]interface{}{
		"secret/": map[string]string{"type": "kv", "description": ""},
		"pki/":    map[string]string{"type": "pki", "description": ""},
	}
	srv := newMockEngineCheckServer(t, mounts)
	defer srv.Close()

	mon := newEngineMonitor(t, srv, []string{"secret/", "pki/"})
	results, err := mon.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if !r.Healthy {
			t.Errorf("expected %s to be healthy", r.Path)
		}
	}
}

func TestEngineMonitor_MissingEngine(t *testing.T) {
	mounts := map[string]interface{}{
		"secret/": map[string]string{"type": "kv", "description": ""},
	}
	srv := newMockEngineCheckServer(t, mounts)
	defer srv.Close()

	mon := newEngineMonitor(t, srv, []string{"secret/", "pki/"})
	results, err := mon.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	unhealthy := 0
	for _, r := range results {
		if !r.Healthy {
			unhealthy++
		}
	}
	if unhealthy != 1 {
		t.Errorf("expected 1 unhealthy, got %d", unhealthy)
	}
}
