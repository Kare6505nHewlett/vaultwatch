package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

func newMockReplicationCheckServer(drMode, drState string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dr":          map[string]string{"mode": drMode, "state": drState},
				"performance": map[string]string{"mode": "disabled", "state": ""},
			},
		})
	}))
}

func newReplicationMonitor(t *testing.T, addr string) *ReplicationMonitor {
	t.Helper()
	client, _ := vault.NewClient(addr, "tok")
	checker, _ := vault.NewReplicationChecker(client, zap.NewNop())
	m, err := NewReplicationMonitor(checker, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return m
}

func TestNewReplicationMonitor_NilChecker(t *testing.T) {
	_, err := NewReplicationMonitor(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewReplicationMonitor_NilLogger(t *testing.T) {
	client, _ := vault.NewClient("http://127.0.0.1:8200", "tok")
	checker, _ := vault.NewReplicationChecker(client, zap.NewNop())
	_, err := NewReplicationMonitor(checker, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReplicationMonitor_Healthy(t *testing.T) {
	server := newMockReplicationCheckServer("primary", "running")
	defer server.Close()

	m := newReplicationMonitor(t, server.URL)
	result, err := m.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Healthy {
		t.Errorf("expected healthy, got unhealthy: %s", result.Message)
	}
}

func TestReplicationMonitor_Unhealthy(t *testing.T) {
	server := newMockReplicationCheckServer("primary", "idle")
	defer server.Close()

	m := newReplicationMonitor(t, server.URL)
	result, err := m.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Healthy {
		t.Error("expected unhealthy result")
	}
}
