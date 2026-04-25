package monitor_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockTokenCountServer(total int, byPolicy map[string]int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"total":     total,
				"by_policy": byPolicy,
			},
		})
	}))
}

func newTokenCountMonitor(t *testing.T, srv *httptest.Server, threshold int) *monitor.TokenCountMonitor {
	t.Helper()
	logger := zap.NewNop()
	client := &vault.Client{
		Address: srv.URL,
		Token:   "test-token",
		HTTP:    srv.Client(),
	}
	checker, err := vault.NewTokenCountChecker(client, logger)
	if err != nil {
		t.Fatalf("NewTokenCountChecker: %v", err)
	}
	mon, err := monitor.NewTokenCountMonitor(checker, threshold, logger)
	if err != nil {
		t.Fatalf("NewTokenCountMonitor: %v", err)
	}
	return mon
}

func TestNewTokenCountMonitor_NilChecker(t *testing.T) {
	_, err := monitor.NewTokenCountMonitor(nil, 100, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewTokenCountMonitor_NilLogger(t *testing.T) {
	srv := newMockTokenCountServer(0, nil)
	defer srv.Close()
	client := &vault.Client{Address: srv.URL, Token: "tok", HTTP: srv.Client()}
	checker, _ := vault.NewTokenCountChecker(client, zap.NewNop())
	_, err := monitor.NewTokenCountMonitor(checker, 100, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestTokenCountMonitor_BelowThreshold(t *testing.T) {
	srv := newMockTokenCountServer(50, map[string]int{"default": 50})
	defer srv.Close()
	mon := newTokenCountMonitor(t, srv, 100)

	res, err := mon.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Exceeded {
		t.Errorf("expected threshold not exceeded")
	}
	if res.Total != 50 {
		t.Errorf("expected total 50, got %d", res.Total)
	}
}

func TestTokenCountMonitor_ExceedsThreshold(t *testing.T) {
	srv := newMockTokenCountServer(200, map[string]int{"default": 200})
	defer srv.Close()
	mon := newTokenCountMonitor(t, srv, 100)

	res, err := mon.Check()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Exceeded {
		t.Errorf("expected threshold exceeded")
	}
	if res.Total != 200 {
		t.Errorf("expected total 200, got %d", res.Total)
	}
}
