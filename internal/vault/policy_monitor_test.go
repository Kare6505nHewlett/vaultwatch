package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

func newPolicyMonitorServer(policies map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/v1/sys/policy/"):]
		if body, ok := policies[name]; ok {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"name": name, "rules": body})
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestNewPolicyMonitor_NilChecker(t *testing.T) {
	logger := zap.NewNop()
	_, err := vault.NewPolicyMonitor(nil, logger, []string{"default"})
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewPolicyMonitor_NilLogger(t *testing.T) {
	svr := newPolicyMonitorServer(nil)
	defer svr.Close()
	client, _ := vault.NewClient(svr.URL, "test-token")
	checker, _ := vault.NewPolicyChecker(client, zap.NewNop())
	_, err := vault.NewPolicyMonitor(checker, nil, []string{"default"})
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestNewPolicyMonitor_NoPolicies(t *testing.T) {
	svr := newPolicyMonitorServer(nil)
	defer svr.Close()
	client, _ := vault.NewClient(svr.URL, "test-token")
	checker, _ := vault.NewPolicyChecker(client, zap.NewNop())
	_, err := vault.NewPolicyMonitor(checker, zap.NewNop(), nil)
	if err == nil {
		t.Fatal("expected error for empty policies")
	}
}

func TestPolicyMonitor_Check_AllExist(t *testing.T) {
	policies := map[string]string{"default": "path \"*\" { capabilities = [\"read\"] }"}
	svr := newPolicyMonitorServer(policies)
	defer svr.Close()
	client, _ := vault.NewClient(svr.URL, "test-token")
	checker, _ := vault.NewPolicyChecker(client, zap.NewNop())
	monitor, _ := vault.NewPolicyMonitor(checker, zap.NewNop(), []string{"default"})
	results := monitor.Check(context.Background())
	if len(results) != 1 || !results[0].Exists {
		t.Errorf("expected policy to exist, got %+v", results)
	}
}

func TestPolicyMonitor_Check_Missing(t *testing.T) {
	svr := newPolicyMonitorServer(map[string]string{})
	defer svr.Close()
	client, _ := vault.NewClient(svr.URL, "test-token")
	checker, _ := vault.NewPolicyChecker(client, zap.NewNop())
	monitor, _ := vault.NewPolicyMonitor(checker, zap.NewNop(), []string{"missing-policy"})
	results := monitor.Check(context.Background())
	if len(results) != 1 || results[0].Exists {
		t.Errorf("expected policy to be missing, got %+v", results)
	}
}
