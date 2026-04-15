package monitor_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockTokenCheckServer(ttlSeconds int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"ttl":           ttlSeconds,
				"display_name":  "test-token",
				"renewable":     true,
			},
		})
	}))
}

func newTokenChecker(t *testing.T, srv *httptest.Server, warn time.Duration) *monitor.TokenChecker {
	t.Helper()
	t.Setenv("VAULT_ADDR", srv.URL)
	t.Setenv("VAULT_TOKEN", "test-token")
	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	checker, err := monitor.NewTokenChecker(client, logger, warn)
	if err != nil {
		t.Fatalf("failed to create token checker: %v", err)
	}
	return checker
}

func TestCheckToken_OK(t *testing.T) {
	srv := newMockTokenCheckServer(3600)
	defer srv.Close()
	checker := newTokenChecker(t, srv, 1*time.Hour)
	result, err := checker.CheckToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != monitor.StatusOK {
		t.Errorf("expected StatusOK, got %v", result.Status)
	}
}

func TestCheckToken_Warning(t *testing.T) {
	srv := newMockTokenCheckServer(1800)
	defer srv.Close()
	checker := newTokenChecker(t, srv, 2*time.Hour)
	result, err := checker.CheckToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != monitor.StatusWarning {
		t.Errorf("expected StatusWarning, got %v", result.Status)
	}
}

func TestCheckToken_Expired(t *testing.T) {
	srv := newMockTokenCheckServer(0)
	defer srv.Close()
	checker := newTokenChecker(t, srv, 1*time.Hour)
	result, err := checker.CheckToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != monitor.StatusExpired {
		t.Errorf("expected StatusExpired, got %v", result.Status)
	}
}

func TestNewTokenChecker_NilClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := monitor.NewTokenChecker(nil, logger, time.Hour)
	if err == nil {
		t.Error("expected error for nil client")
	}
}

func TestNewTokenChecker_NilLogger(t *testing.T) {
	srv := newMockTokenCheckServer(3600)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "token")
	_, err := monitor.NewTokenChecker(client, nil, time.Hour)
	if err == nil {
		t.Error("expected error for nil logger")
	}
}
