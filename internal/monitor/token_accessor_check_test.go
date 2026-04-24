package monitor_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap/zaptest"
)

func newMockTokenAccessorCheckServer(t *testing.T, ttl int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"accessor":     "testacc",
				"display_name": "test",
				"policies":     []string{"default"},
				"ttl":          ttl,
				"expire_time":  "2099-01-01T00:00:00Z",
			},
		})
	}))
}

func newTokenAccessorMonitor(t *testing.T, serverURL string, accessors []string, warnTTL time.Duration) *monitor.TokenAccessorMonitor {
	t.Helper()
	logger := zaptest.NewLogger(t)
	client, err := vault.NewClient(serverURL, "tok")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewTokenAccessorChecker(client, logger)
	if err != nil {
		t.Fatalf("NewTokenAccessorChecker: %v", err)
	}
	m, err := monitor.NewTokenAccessorMonitor(checker, accessors, warnTTL, logger)
	if err != nil {
		t.Fatalf("NewTokenAccessorMonitor: %v", err)
	}
	return m
}

func TestNewTokenAccessorMonitor_NilChecker(t *testing.T) {
	logger := zaptest.NewLogger(t)
	_, err := monitor.NewTokenAccessorMonitor(nil, nil, time.Hour, logger)
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewTokenAccessorMonitor_NilLogger(t *testing.T) {
	server := newMockTokenAccessorCheckServer(t, 3600)
	defer server.Close()
	client, _ := vault.NewClient(server.URL, "tok")
	logger := zaptest.NewLogger(t)
	checker, _ := vault.NewTokenAccessorChecker(client, logger)
	_, err := monitor.NewTokenAccessorMonitor(checker, nil, time.Hour, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestTokenAccessorMonitor_OK(t *testing.T) {
	server := newMockTokenAccessorCheckServer(t, 7200)
	defer server.Close()
	m := newTokenAccessorMonitor(t, server.URL, []string{"testacc"}, time.Hour)
	results := m.Check()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "ok" {
		t.Errorf("expected ok, got %s", results[0].Status)
	}
}

func TestTokenAccessorMonitor_Warning(t *testing.T) {
	server := newMockTokenAccessorCheckServer(t, 1800)
	defer server.Close()
	m := newTokenAccessorMonitor(t, server.URL, []string{"testacc"}, 2*time.Hour)
	results := m.Check()
	if results[0].Status != "warning" {
		t.Errorf("expected warning, got %s", results[0].Status)
	}
}

func TestTokenAccessorMonitor_Expired(t *testing.T) {
	server := newMockTokenAccessorCheckServer(t, 0)
	defer server.Close()
	m := newTokenAccessorMonitor(t, server.URL, []string{"testacc"}, time.Hour)
	results := m.Check()
	if results[0].Status != "expired" {
		t.Errorf("expected expired, got %s", results[0].Status)
	}
}
