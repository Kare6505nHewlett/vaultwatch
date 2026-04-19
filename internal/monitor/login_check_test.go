package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

func newMockLoginCheckServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func newLoginMonitor(t *testing.T, addr string) *AppRoleLoginMonitor {
	t.Helper()
	client, _ := vault.NewClient(addr, "tok")
	logger, _ := zap.NewDevelopment()
	checker, err := vault.NewLoginChecker(client, logger)
	if err != nil {
		t.Fatalf("NewLoginChecker: %v", err)
	}
	mon, err := NewAppRoleLoginMonitor(checker, "role", "secret", logger)
	if err != nil {
		t.Fatalf("NewAppRoleLoginMonitor: %v", err)
	}
	return mon
}

func TestNewAppRoleLoginMonitor_NilChecker(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := NewAppRoleLoginMonitor(nil, "r", "s", logger)
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewAppRoleLoginMonitor_NilLogger(t *testing.T) {
	client, _ := vault.NewClient("http://127.0.0.1:8200", "tok")
	logger, _ := zap.NewDevelopment()
	checker, _ := vault.NewLoginChecker(client, logger)
	_, err := NewAppRoleLoginMonitor(checker, "r", "s", nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestLoginMonitor_Success(t *testing.T) {
	body := map[string]interface{}{
		"auth": map[string]interface{}{
			"client_token":   "s.abc",
			"accessor":       "acc",
			"lease_duration": 1800,
			"renewable":      true,
		},
	}
	srv := newMockLoginCheckServer(t, http.StatusOK, body)
	defer srv.Close()

	mon := newLoginMonitor(t, srv.URL)
	res := mon.Check()
	if !res.Success {
		t.Errorf("expected success, got: %s", res.Message)
	}
	if res.TTL != 1800 {
		t.Errorf("expected TTL 1800, got %d", res.TTL)
	}
}

func TestLoginMonitor_Failure(t *testing.T) {
	srv := newMockLoginCheckServer(t, http.StatusForbidden, nil)
	defer srv.Close()

	mon := newLoginMonitor(t, srv.URL)
	res := mon.Check()
	if res.Success {
		t.Error("expected failure result")
	}
}
