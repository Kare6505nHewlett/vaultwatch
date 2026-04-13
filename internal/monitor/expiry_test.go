package monitor_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockVaultServer(leaseDuration int, renewable bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		resp := map[string]interface{}{
			"lease_id":       "test-lease-id",
			"lease_duration": leaseDuration,
			"renewable":      renewable,
			"data":           map[string]string{"key": "value"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func newTestChecker(t *testing.T, server *httptest.Server, warnThreshold time.Duration) *monitor.Checker {
	t.Helper()
	client, err := vault.NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	return monitor.NewChecker(client, warnThreshold)
}

func TestCheckSecret_Valid(t *testing.T) {
	server := newMockVaultServer(3600, true)
	defer server.Close()

	checker := newTestChecker(t, server, 10*time.Minute)
	status, err := checker.CheckSecret("secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Warning || status.Expired {
		t.Errorf("expected valid status, got warning=%v expired=%v", status.Warning, status.Expired)
	}
}

func TestCheckSecret_Warning(t *testing.T) {
	server := newMockVaultServer(300, true) // 5 minutes lease
	defer server.Close()

	checker := newTestChecker(t, server, 10*time.Minute) // warn if < 10 minutes
	status, err := checker.CheckSecret("secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Warning {
		t.Errorf("expected warning status, got none")
	}
}

func TestCheckSecret_Expired(t *testing.T) {
	server := newMockVaultServer(0, false)
	defer server.Close()

	checker := newTestChecker(t, server, 10*time.Minute)
	status, err := checker.CheckSecret("secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Expired {
		t.Errorf("expected expired status, got none")
	}
}

func TestCheckSecrets_Multiple(t *testing.T) {
	server := newMockVaultServer(3600, true)
	defer server.Close()

	checker := newTestChecker(t, server, 10*time.Minute)
	paths := []string{"secret/data/app1", "secret/data/app2"}
	statuses, err := checker.CheckSecrets(paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(statuses))
	}
}
