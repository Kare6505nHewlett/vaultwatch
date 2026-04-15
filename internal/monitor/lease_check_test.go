package monitor_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

func newMockLeaseCheckServer(t *testing.T, ttl int, renewable bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"ttl":       ttl,
				"renewable": renewable,
				"expire_time": time.Now().Add(time.Duration(ttl) * time.Second).Format(time.RFC3339),
			},
		})
	}))
}

func newLeaseChecker(t *testing.T, serverURL string) *monitor.LeaseChecker {
	t.Helper()
	client, err := vault.NewClient(serverURL, "test-token")
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	logger, _ := zap.NewDevelopment()
	checker, err := monitor.NewLeaseChecker(client, logger)
	if err != nil {
		t.Fatalf("failed to create LeaseChecker: %v", err)
	}
	return checker
}

func TestLeaseCheck_OK(t *testing.T) {
	server := newMockLeaseCheckServer(t, 7200, true)
	defer server.Close()

	checker := newLeaseChecker(t, server.URL)
	result, err := checker.Check("lease/abc123", 1*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != monitor.StatusOK {
		t.Errorf("expected OK, got %s", result.Status)
	}
	if result.LeaseID != "lease/abc123" {
		t.Errorf("expected lease ID lease/abc123, got %s", result.LeaseID)
	}
}

func TestLeaseCheck_Warning(t *testing.T) {
	server := newMockLeaseCheckServer(t, 1800, true)
	defer server.Close()

	checker := newLeaseChecker(t, server.URL)
	result, err := checker.Check("lease/abc123", 1*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != monitor.StatusWarning {
		t.Errorf("expected Warning, got %s", result.Status)
	}
}

func TestLeaseCheck_Expired(t *testing.T) {
	server := newMockLeaseCheckServer(t, 0, false)
	defer server.Close()

	checker := newLeaseChecker(t, server.URL)
	result, err := checker.Check("lease/abc123", 1*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != monitor.StatusExpired {
		t.Errorf("expected Expired, got %s", result.Status)
	}
}

func TestLeaseCheck_EmptyLeaseID(t *testing.T) {
	server := newMockLeaseCheckServer(t, 3600, true)
	defer server.Close()

	checker := newLeaseChecker(t, server.URL)
	_, err := checker.Check("", 1*time.Hour)
	if err == nil {
		t.Error("expected error for empty lease ID, got nil")
	}
}
