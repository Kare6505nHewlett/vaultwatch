package monitor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

func newMockCapabilitiesCheckServer(t *testing.T, capabilities []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"capabilities": capabilities},
		})
	}))
}

func newCapabilitiesMonitor(t *testing.T, serverURL string, reqs []CapabilityRequirement) *CapabilitiesMonitor {
	t.Helper()
	client, err := vault.NewClient(serverURL, "test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	checker, err := vault.NewCapabilitiesChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("failed to create checker: %v", err)
	}
	mon, err := NewCapabilitiesMonitor(checker, reqs, zap.NewNop())
	if err != nil {
		t.Fatalf("failed to create monitor: %v", err)
	}
	return mon
}

func TestNewCapabilitiesMonitor_NilChecker(t *testing.T) {
	_, err := NewCapabilitiesMonitor(nil, nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewCapabilitiesMonitor_NilLogger(t *testing.T) {
	server := newMockCapabilitiesCheckServer(t, []string{"read"})
	defer server.Close()
	client, _ := vault.NewClient(server.URL, "token")
	checker, _ := vault.NewCapabilitiesChecker(client, zap.NewNop())
	_, err := NewCapabilitiesMonitor(checker, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestCapabilitiesMonitor_AllPresent(t *testing.T) {
	server := newMockCapabilitiesCheckServer(t, []string{"read", "list"})
	defer server.Close()

	reqs := []CapabilityRequirement{
		{Path: "secret/data/app", RequiredCapabilities: []string{"read", "list"}},
	}
	mon := newCapabilitiesMonitor(t, server.URL, reqs)
	results, err := mon.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Healthy {
		t.Errorf("expected healthy result, got missing: %v", results[0].Missing)
	}
}

func TestCapabilitiesMonitor_MissingCapability(t *testing.T) {
	server := newMockCapabilitiesCheckServer(t, []string{"read"})
	defer server.Close()

	reqs := []CapabilityRequirement{
		{Path: "secret/data/app", RequiredCapabilities: []string{"read", "write"}},
	}
	mon := newCapabilitiesMonitor(t, server.URL, reqs)
	results, err := mon.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Healthy {
		t.Error("expected unhealthy result due to missing write capability")
	}
	if len(results[0].Missing) != 1 || results[0].Missing[0] != "write" {
		t.Errorf("expected missing [write], got %v", results[0].Missing)
	}
}
