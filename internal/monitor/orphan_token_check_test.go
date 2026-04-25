package monitor

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockOrphanCheckServer(t *testing.T, orphan bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":       "tok-abc",
				"orphan":   orphan,
				"policies": []string{"default"},
				"ttl":      1800,
			},
		})
	}))
}

func newOrphanMonitor(t *testing.T, srv *httptest.Server, tokens []string) *OrphanTokenMonitor {
	t.Helper()
	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewOrphanTokenChecker(client, log.New(os.Stderr, "", 0))
	if err != nil {
		t.Fatalf("NewOrphanTokenChecker: %v", err)
	}
	mon, err := NewOrphanTokenMonitor(checker, tokens, log.New(os.Stderr, "", 0))
	if err != nil {
		t.Fatalf("NewOrphanTokenMonitor: %v", err)
	}
	return mon
}

func TestNewOrphanTokenMonitor_NilChecker(t *testing.T) {
	_, err := NewOrphanTokenMonitor(nil, nil, log.New(os.Stderr, "", 0))
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewOrphanTokenMonitor_NilLogger(t *testing.T) {
	srv := newMockOrphanCheckServer(t, true)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "tok")
	checker, _ := vault.NewOrphanTokenChecker(client, log.New(os.Stderr, "", 0))
	_, err := NewOrphanTokenMonitor(checker, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestOrphanTokenMonitor_IsOrphan(t *testing.T) {
	srv := newMockOrphanCheckServer(t, true)
	defer srv.Close()

	mon := newOrphanMonitor(t, srv, []string{"tok-abc"})
	results := mon.CheckAll()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Orphan {
		t.Errorf("expected orphan=true")
	}
	if results[0].TTL != 1800 {
		t.Errorf("expected TTL=1800, got %d", results[0].TTL)
	}
}

func TestOrphanTokenMonitor_NotOrphan(t *testing.T) {
	srv := newMockOrphanCheckServer(t, false)
	defer srv.Close()

	mon := newOrphanMonitor(t, srv, []string{"tok-xyz"})
	results := mon.CheckAll()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Orphan {
		t.Errorf("expected orphan=false")
	}
}

func TestOrphanTokenMonitor_NoTokens(t *testing.T) {
	srv := newMockOrphanCheckServer(t, true)
	defer srv.Close()

	mon := newOrphanMonitor(t, srv, []string{})
	results := mon.CheckAll()
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
