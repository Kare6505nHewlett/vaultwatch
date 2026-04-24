package vault

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newMockTransitServer(t *testing.T, keyName string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"type":                    "aes256-gcm96",
				"deletion_allowed":        false,
				"exportable":              true,
				"latest_version":          3,
				"min_decryption_version":  1,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func newTransitChecker(t *testing.T, addr string) *TransitChecker {
	t.Helper()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewTransitChecker(client, log.New(os.Stderr, "", 0))
	if err != nil {
		t.Fatalf("NewTransitChecker: %v", err)
	}
	return checker
}

func TestNewTransitChecker_NilClient(t *testing.T) {
	_, err := NewTransitChecker(nil, log.New(os.Stderr, "", 0))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTransitChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://localhost:8200", "tok")
	_, err := NewTransitChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetTransitKeyInfo_Success(t *testing.T) {
	svr := newMockTransitServer(t, "my-key", http.StatusOK)
	defer svr.Close()

	checker := newTransitChecker(t, svr.URL)
	info, err := checker.GetTransitKeyInfo("my-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "my-key" {
		t.Errorf("expected name 'my-key', got %q", info.Name)
	}
	if info.Type != "aes256-gcm96" {
		t.Errorf("expected type 'aes256-gcm96', got %q", info.Type)
	}
	if info.LatestVersion != 3 {
		t.Errorf("expected latest version 3, got %d", info.LatestVersion)
	}
}

func TestGetTransitKeyInfo_NotFound(t *testing.T) {
	svr := newMockTransitServer(t, "missing-key", http.StatusNotFound)
	defer svr.Close()

	checker := newTransitChecker(t, svr.URL)
	_, err := checker.GetTransitKeyInfo("missing-key")
	if err == nil {
		t.Fatal("expected error for not found key")
	}
}

func TestGetTransitKeyInfo_EmptyName(t *testing.T) {
	svr := newMockTransitServer(t, "", http.StatusOK)
	defer svr.Close()

	checker := newTransitChecker(t, svr.URL)
	_, err := checker.GetTransitKeyInfo("")
	if err == nil {
		t.Fatal("expected error for empty key name")
	}
}
