package monitor_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockAuditCheckServer(t *testing.T, payload map[string]any, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

func newAuditMonitor(t *testing.T, serverURL string) *monitor.AuditMonitor {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	client, err := vault.NewClient(serverURL, "test-token", logger)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	checker := vault.NewAuditChecker(client, logger)
	m, err := monitor.NewAuditMonitor(checker, logger)
	if err != nil {
		t.Fatalf("failed to create audit monitor: %v", err)
	}
	return m
}

func TestNewAuditMonitor_NilChecker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err := monitor.NewAuditMonitor(nil, logger)
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewAuditMonitor_NilLogger(t *testing.T) {
	server := newMockAuditCheckServer(t, map[string]any{"data": map[string]any{}}, http.StatusOK)
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	client, _ := vault.NewClient(server.URL, "test-token", logger)
	checker := vault.NewAuditChecker(client, logger)

	_, err := monitor.NewAuditMonitor(checker, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestAuditMonitor_DevicesPresent(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"file/": map[string]any{
				"type":        "file",
				"description": "file audit device",
			},
		},
	}
	server := newMockAuditCheckServer(t, payload, http.StatusOK)
	defer server.Close()

	m := newAuditMonitor(t, server.URL)
	result := m.Check()

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.DeviceCount < 1 {
		t.Errorf("expected at least 1 device, got %d", result.DeviceCount)
	}
	if !result.Healthy {
		t.Error("expected healthy result when devices are present")
	}
}

func TestAuditMonitor_NoDevices(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{},
	}
	server := newMockAuditCheckServer(t, payload, http.StatusOK)
	defer server.Close()

	m := newAuditMonitor(t, server.URL)
	result := m.Check()

	if result.Healthy {
		t.Error("expected unhealthy result when no audit devices are present")
	}
	if result.DeviceCount != 0 {
		t.Errorf("expected 0 devices, got %d", result.DeviceCount)
	}
}

func TestAuditMonitor_ServerError(t *testing.T) {
	server := newMockAuditCheckServer(t, map[string]any{}, http.StatusInternalServerError)
	defer server.Close()

	m := newAuditMonitor(t, server.URL)
	result := m.Check()

	if result.Error == nil {
		t.Fatal("expected error on server failure")
	}
	if result.Healthy {
		t.Error("expected unhealthy result on server error")
	}
}
