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

func newMockTransitCheckServer(t *testing.T, found bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"type":                   "aes256-gcm96",
				"deletion_allowed":       false,
				"exportable":             false,
				"latest_version":         2,
				"min_decryption_version": 1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func newTransitMonitor(t *testing.T, addr string, keys []string) *TransitMonitor {
	t.Helper()
	client, err := vault.NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewTransitChecker(client, log.New(os.Stderr, "", 0))
	if err != nil {
		t.Fatalf("NewTransitChecker: %v", err)
	}
	monitor, err := NewTransitMonitor(checker, keys, log.New(os.Stderr, "", 0))
	if err != nil {
		t.Fatalf("NewTransitMonitor: %v", err)
	}
	return monitor
}

func TestNewTransitMonitor_NilChecker(t *testing.T) {
	_, err := NewTransitMonitor(nil, []string{"key"}, log.New(os.Stderr, "", 0))
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewTransitMonitor_NilLogger(t *testing.T) {
	svr := newMockTransitCheckServer(t, true)
	defer svr.Close()
	client, _ := vault.NewClient(svr.URL, "tok")
	checker, _ := vault.NewTransitChecker(client, log.New(os.Stderr, "", 0))
	_, err := NewTransitMonitor(checker, []string{"key"}, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestTransitMonitor_AllPresent(t *testing.T) {
	svr := newMockTransitCheckServer(t, true)
	defer svr.Close()

	m := newTransitMonitor(t, svr.URL, []string{"key-a", "key-b"})
	results := m.Check()

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Healthy {
			t.Errorf("expected key %q to be healthy", r.KeyName)
		}
	}
}

func TestTransitMonitor_KeyMissing(t *testing.T) {
	svr := newMockTransitCheckServer(t, false)
	defer svr.Close()

	m := newTransitMonitor(t, svr.URL, []string{"missing-key"})
	results := m.Check()

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Healthy {
		t.Error("expected key to be unhealthy")
	}
}
