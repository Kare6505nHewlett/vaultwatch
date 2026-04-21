package monitor_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockPluginCheckServer(keys []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": keys},
		})
	}))
}

func newPluginMonitor(t *testing.T, srv *httptest.Server, pluginType string, expected []string) *monitor.PluginMonitor {
	t.Helper()
	logger := zap.NewNop()
	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewPluginChecker(client, logger)
	if err != nil {
		t.Fatalf("NewPluginChecker: %v", err)
	}
	m, err := monitor.NewPluginMonitor(checker, logger, pluginType, expected)
	if err != nil {
		t.Fatalf("NewPluginMonitor: %v", err)
	}
	return m
}

func TestNewPluginMonitor_NilChecker(t *testing.T) {
	_, err := monitor.NewPluginMonitor(nil, zap.NewNop(), "secret", nil)
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewPluginMonitor_NilLogger(t *testing.T) {
	srv := newMockPluginCheckServer([]string{})
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "tok")
	checker, _ := vault.NewPluginChecker(client, zap.NewNop())
	_, err := monitor.NewPluginMonitor(checker, nil, "secret", nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestPluginMonitor_AllPresent(t *testing.T) {
	srv := newMockPluginCheckServer([]string{"aws", "gcp", "azure"})
	defer srv.Close()

	m := newPluginMonitor(t, srv, "secret", []string{"aws", "gcp"})
	result, err := m.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !result.Healthy {
		t.Errorf("expected healthy, got missing: %v", result.Missing)
	}
	if len(result.Present) != 2 {
		t.Errorf("expected 2 present, got %d", len(result.Present))
	}
}

func TestPluginMonitor_SomeMissing(t *testing.T) {
	srv := newMockPluginCheckServer([]string{"aws"})
	defer srv.Close()

	m := newPluginMonitor(t, srv, "secret", []string{"aws", "gcp"})
	result, err := m.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if result.Healthy {
		t.Error("expected unhealthy due to missing plugin")
	}
	if len(result.Missing) != 1 || result.Missing[0] != "gcp" {
		t.Errorf("expected [gcp] missing, got %v", result.Missing)
	}
}
