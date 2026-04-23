package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockTokenMetadataCheckServer(t *testing.T, policies []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"display_name": "monitor-test-token",
				"policies":    policies,
				"entity_id":   "entity-xyz",
				"orphan":      false,
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

func newTokenMetadataMonitor(t *testing.T, addr, accessor string, required []string) *TokenMetadataMonitor {
	t.Helper()
	client, err := vault.NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewTokenMetadataChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTokenMetadataChecker: %v", err)
	}
	mon, err := NewTokenMetadataMonitor(checker, accessor, required, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTokenMetadataMonitor: %v", err)
	}
	return mon
}

func TestNewTokenMetadataMonitor_NilChecker(t *testing.T) {
	_, err := NewTokenMetadataMonitor(nil, "acc", nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewTokenMetadataMonitor_NilLogger(t *testing.T) {
	client, _ := vault.NewClient("http://localhost:8200", "tok")
	checker, _ := vault.NewTokenMetadataChecker(client, zap.NewNop())
	_, err := NewTokenMetadataMonitor(checker, "acc", nil, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestTokenMetadataMonitor_AllPoliciesPresent(t *testing.T) {
	server := newMockTokenMetadataCheckServer(t, []string{"default", "admin", "read-only"})
	defer server.Close()

	mon := newTokenMetadataMonitor(t, server.URL, "test-acc", []string{"default", "admin"})
	result, err := mon.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Healthy {
		t.Errorf("expected healthy, got message: %s", result.Message)
	}
	if len(result.MissingPolicies) != 0 {
		t.Errorf("expected no missing policies, got %v", result.MissingPolicies)
	}
}

func TestTokenMetadataMonitor_MissingPolicy(t *testing.T) {
	server := newMockTokenMetadataCheckServer(t, []string{"default"})
	defer server.Close()

	mon := newTokenMetadataMonitor(t, server.URL, "test-acc", []string{"default", "superadmin"})
	result, err := mon.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Healthy {
		t.Error("expected unhealthy due to missing policy")
	}
	if len(result.MissingPolicies) != 1 || result.MissingPolicies[0] != "superadmin" {
		t.Errorf("unexpected missing policies: %v", result.MissingPolicies)
	}
}
