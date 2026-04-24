package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

func newMockTokenListMonitorServer(keys []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodList {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": keys},
		})
	}))
}

func newTokenListMonitor(t *testing.T, srv *httptest.Server, maxTokens int) *vault.TokenListMonitor {
	t.Helper()
	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	lister, err := vault.NewTokenLister(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTokenLister: %v", err)
	}
	mon, err := vault.NewTokenListMonitor(lister, zap.NewNop(), maxTokens)
	if err != nil {
		t.Fatalf("NewTokenListMonitor: %v", err)
	}
	return mon
}

func TestNewTokenListMonitor_NilLister(t *testing.T) {
	_, err := vault.NewTokenListMonitor(nil, zap.NewNop(), 0)
	if err == nil {
		t.Fatal("expected error for nil lister")
	}
}

func TestNewTokenListMonitor_NilLogger(t *testing.T) {
	srv := newMockTokenListMonitorServer(nil)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "tok")
	lister, _ := vault.NewTokenLister(client, zap.NewNop())
	_, err := vault.NewTokenListMonitor(lister, nil, 0)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestTokenListMonitor_BelowThreshold(t *testing.T) {
	srv := newMockTokenListMonitorServer([]string{"tok1", "tok2"})
	defer srv.Close()
	mon := newTokenListMonitor(t, srv, 5)
	res, err := mon.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Warning {
		t.Errorf("expected no warning, got warning: %s", res.Message)
	}
	if res.Count != 2 {
		t.Errorf("expected count 2, got %d", res.Count)
	}
}

func TestTokenListMonitor_ExceedsThreshold(t *testing.T) {
	srv := newMockTokenListMonitorServer([]string{"tok1", "tok2", "tok3"})
	defer srv.Close()
	mon := newTokenListMonitor(t, srv, 2)
	res, err := mon.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Warning {
		t.Errorf("expected warning, got none")
	}
	if res.Count != 3 {
		t.Errorf("expected count 3, got %d", res.Count)
	}
}
