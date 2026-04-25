package monitor_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockTokenExpireServer(ttl int, renewable bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expireTime := time.Now().Add(time.Duration(ttl) * time.Second).Format(time.RFC3339)
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"id":          "test-token-id",
				"ttl":         ttl,
				"renewable":   renewable,
				"expire_time": expireTime,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func newTokenExpireMonitor(t *testing.T, server *httptest.Server, warn time.Duration) *monitor.TokenExpireMonitor {
	t.Helper()
	logger := zap.NewNop()
	client, err := vault.NewClient(server.URL, "test-token", logger)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewTokenExpireChecker(client, logger)
	if err != nil {
		t.Fatalf("NewTokenExpireChecker: %v", err)
	}
	m, err := monitor.NewTokenExpireMonitor(checker, "test-token", warn, logger)
	if err != nil {
		t.Fatalf("NewTokenExpireMonitor: %v", err)
	}
	return m
}

func TestNewTokenExpireMonitor_NilChecker(t *testing.T) {
	_, err := monitor.NewTokenExpireMonitor(nil, "tok", time.Hour, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewTokenExpireMonitor_NilLogger(t *testing.T) {
	server := newMockTokenExpireServer(3600, true)
	defer server.Close()
	client, _ := vault.NewClient(server.URL, "tok", zap.NewNop())
	checker, _ := vault.NewTokenExpireChecker(client, zap.NewNop())
	_, err := monitor.NewTokenExpireMonitor(checker, "tok", time.Hour, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestTokenExpireMonitor_OK(t *testing.T) {
	server := newMockTokenExpireServer(7200, true)
	defer server.Close()
	m := newTokenExpireMonitor(t, server, time.Hour)
	result := m.Check(context.Background())
	if result.Status != monitor.StatusOK {
		t.Errorf("expected OK, got %s: %s", result.Status, result.Message)
	}
}

func TestTokenExpireMonitor_Warning(t *testing.T) {
	server := newMockTokenExpireServer(1800, true)
	defer server.Close()
	m := newTokenExpireMonitor(t, server, 2*time.Hour)
	result := m.Check(context.Background())
	if result.Status != monitor.StatusWarning {
		t.Errorf("expected Warning, got %s: %s", result.Status, result.Message)
	}
}

func TestTokenExpireMonitor_Expired(t *testing.T) {
	server := newMockTokenExpireServer(-1, false)
	defer server.Close()
	m := newTokenExpireMonitor(t, server, time.Hour)
	result := m.Check(context.Background())
	if result.Status != monitor.StatusExpired {
		t.Errorf("expected Expired, got %s: %s", result.Status, result.Message)
	}
}
