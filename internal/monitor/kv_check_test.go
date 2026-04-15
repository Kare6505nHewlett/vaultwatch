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

func newMockKVServer(t *testing.T, metadata map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": metadata})
	}))
}

func newKVChecker(t *testing.T, serverURL string) *monitor.KVChecker {
	t.Helper()
	client, err := vault.NewClient(serverURL, "test-token")
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	logger, _ := zap.NewDevelopment()
	checker, err := monitor.NewKVChecker(client, logger)
	if err != nil {
		t.Fatalf("failed to create KVChecker: %v", err)
	}
	return checker
}

func TestKVCheck_SecretNotExpired(t *testing.T) {
	futureTime := time.Now().Add(72 * time.Hour).Format(time.RFC3339)
	server := newMockKVServer(t, map[string]interface{}{
		"expiry": futureTime,
	})
	defer server.Close()

	checker := newKVChecker(t, server.URL)
	result, err := checker.Check("secret/data/myapp", 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != monitor.StatusOK {
		t.Errorf("expected OK, got %s", result.Status)
	}
}

func TestKVCheck_SecretWarning(t *testing.T) {
	soonTime := time.Now().Add(12 * time.Hour).Format(time.RFC3339)
	server := newMockKVServer(t, map[string]interface{}{
		"expiry": soonTime,
	})
	defer server.Close()

	checker := newKVChecker(t, server.URL)
	result, err := checker.Check("secret/data/myapp", 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != monitor.StatusWarning {
		t.Errorf("expected Warning, got %s", result.Status)
	}
}

func TestKVCheck_SecretExpired(t *testing.T) {
	pastTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	server := newMockKVServer(t, map[string]interface{}{
		"expiry": pastTime,
	})
	defer server.Close()

	checker := newKVChecker(t, server.URL)
	result, err := checker.Check("secret/data/myapp", 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != monitor.StatusExpired {
		t.Errorf("expected Expired, got %s", result.Status)
	}
}

func TestKVCheck_NoExpiryField(t *testing.T) {
	server := newMockKVServer(t, map[string]interface{}{
		"value": "no-expiry-here",
	})
	defer server.Close()

	checker := newKVChecker(t, server.URL)
	_, err := checker.Check("secret/data/myapp", 24*time.Hour)
	if err == nil {
		t.Error("expected error for missing expiry field, got nil")
	}
}
