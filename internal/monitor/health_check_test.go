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

func newMockHealthCheckServer(t *testing.T, body map[string]interface{}, code int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(body)
	}))
}

func newHealthMonitor(t *testing.T, serverURL string) *HealthMonitor {
	t.Helper()
	client, err := vault.NewClient(serverURL, "test-token")
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}
	checker, err := vault.NewHealthChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("creating checker: %v", err)
	}
	mon, err := NewHealthMonitor(checker, zap.NewNop())
	if err != nil {
		t.Fatalf("creating monitor: %v", err)
	}
	return mon
}

func TestHealthMonitor_NilChecker(t *testing.T) {
	_, err := NewHealthMonitor(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestHealthMonitor_Healthy(t *testing.T) {
	server := newMockHealthCheckServer(t, map[string]interface{}{
		"initialized": true, "sealed": false, "standby": false,
		"version": "1.15.0", "cluster_name": "primary",
	}, http.StatusOK)
	defer server.Close()

	mon := newHealthMonitor(t, server.URL)
	result := mon.Run(context.Background())

	if !result.IsHealthy() {
		t.Errorf("expected healthy result, got err=%v sealed=%v initialized=%v", result.Err, result.Sealed, result.Initialized)
	}
	if result.Version != "1.15.0" {
		t.Errorf("expected version 1.15.0, got %s", result.Version)
	}
}

func TestHealthMonitor_Sealed(t *testing.T) {
	server := newMockHealthCheckServer(t, map[string]interface{}{
		"initialized": true, "sealed": true, "standby": false, "version": "1.15.0",
	}, http.StatusOK)
	defer server.Close()

	mon := newHealthMonitor(t, server.URL)
	result := mon.Run(context.Background())

	if result.IsHealthy() {
		t.Error("expected unhealthy result for sealed vault")
	}
}

func TestHealthMonitor_RequestError(t *testing.T) {
	mon := newHealthMonitor(t, "http://127.0.0.1:19999")
	result := mon.Run(context.Background())

	if result.Err == nil {
		t.Error("expected error for unreachable server")
	}
	if result.IsHealthy() {
		t.Error("expected unhealthy result on request error")
	}
}
