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

func newMockQuotaServer(t *testing.T, keys []string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": keys},
		})
	}))
}

func newQuotaMonitor(t *testing.T, srv *httptest.Server) *monitor.QuotaMonitor {
	t.Helper()
	logger := zap.NewNop()
	client, err := vault.NewClient(srv.URL, "test-token", logger)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewQuotaChecker(client, logger)
	if err != nil {
		t.Fatalf("NewQuotaChecker: %v", err)
	}
	m, err := monitor.NewQuotaMonitor(checker, logger)
	if err != nil {
		t.Fatalf("NewQuotaMonitor: %v", err)
	}
	return m
}

func TestNewQuotaMonitor_NilChecker(t *testing.T) {
	_, err := monitor.NewQuotaMonitor(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewQuotaMonitor_NilLogger(t *testing.T) {
	srv := newMockQuotaServer(t, nil, http.StatusOK)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "tok", zap.NewNop())
	checker, _ := vault.NewQuotaChecker(client, zap.NewNop())
	_, err := monitor.NewQuotaMonitor(checker, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestQuotaMonitor_WithQuotas(t *testing.T) {
	srv := newMockQuotaServer(t, []string{"global-rate", "api-rate"}, http.StatusOK)
	defer srv.Close()
	m := newQuotaMonitor(t, srv)
	res, err := m.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if res.TotalQuotas != 2 {
		t.Errorf("expected 2 quotas, got %d", res.TotalQuotas)
	}
	if !res.Healthy {
		t.Error("expected healthy result")
	}
}

func TestQuotaMonitor_Empty(t *testing.T) {
	srv := newMockQuotaServer(t, []string{}, http.StatusOK)
	defer srv.Close()
	m := newQuotaMonitor(t, srv)
	res, err := m.Check(context.Background())
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if res.TotalQuotas != 0 {
		t.Errorf("expected 0 quotas, got %d", res.TotalQuotas)
	}
}
