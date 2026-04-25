package monitor

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/your-org/vaultwatch/internal/vault"
)

func newMockTokenRolesCheckServer(t *testing.T, roles map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		for name, data := range roles {
			if r.URL.Path == "/v1/auth/token/roles/"+name {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newTokenRolesMonitor(t *testing.T, addr string, roles []string) *TokenRolesMonitor {
	t.Helper()
	client, _ := vault.NewClient(addr, "test-token")
	logger := log.New(os.Stdout, "", 0)
	checker, err := vault.NewTokenRolesChecker(client, logger)
	if err != nil {
		t.Fatalf("checker error: %v", err)
	}
	monitor, err := NewTokenRolesMonitor(checker, roles, logger)
	if err != nil {
		t.Fatalf("monitor error: %v", err)
	}
	return monitor
}

func TestNewTokenRolesMonitor_NilChecker(t *testing.T) {
	_, err := NewTokenRolesMonitor(nil, nil, log.New(os.Stdout, "", 0))
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewTokenRolesMonitor_NilLogger(t *testing.T) {
	client, _ := vault.NewClient("http://127.0.0.1:8200", "tok")
	checker, _ := vault.NewTokenRolesChecker(client, log.New(os.Stdout, "", 0))
	_, err := NewTokenRolesMonitor(checker, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestTokenRolesMonitor_AllPresent(t *testing.T) {
	roles := map[string]interface{}{
		"app-role": map[string]interface{}{"orphan": true, "renewable": true, "token_max_ttl": 3600, "explicit_max_ttl": 0},
	}
	server := newMockTokenRolesCheckServer(t, roles)
	defer server.Close()

	m := newTokenRolesMonitor(t, server.URL, []string{"app-role"})
	results := m.Check()

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Healthy {
		t.Errorf("expected healthy result for app-role")
	}
	if results[0].MaxTTL != 3600 {
		t.Errorf("expected MaxTTL=3600, got %d", results[0].MaxTTL)
	}
}

func TestTokenRolesMonitor_MissingRole(t *testing.T) {
	server := newMockTokenRolesCheckServer(t, map[string]interface{}{})
	defer server.Close()

	m := newTokenRolesMonitor(t, server.URL, []string{"missing-role"})
	results := m.Check()

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Healthy {
		t.Error("expected unhealthy result for missing role")
	}
	if results[0].Warning == "" {
		t.Error("expected warning message")
	}
}

func TestTokenRolesMonitor_UnlimitedTTLWarning(t *testing.T) {
	roles := map[string]interface{}{
		"no-ttl-role": map[string]interface{}{"orphan": false, "renewable": true, "token_max_ttl": 0, "explicit_max_ttl": 0},
	}
	server := newMockTokenRolesCheckServer(t, roles)
	defer server.Close()

	m := newTokenRolesMonitor(t, server.URL, []string{"no-ttl-role"})
	results := m.Check()

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Healthy {
		t.Error("expected healthy result despite warning")
	}
	if results[0].Warning == "" {
		t.Error("expected warning for unlimited TTL")
	}
}
