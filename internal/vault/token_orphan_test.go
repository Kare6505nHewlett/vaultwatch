package vault

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newMockOrphanServer(t *testing.T, orphan bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/token/lookup" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":       "test-token",
				"orphan":   orphan,
				"policies": []string{"default"},
				"ttl":      3600,
			},
		})
	}))
}

func newOrphanChecker(t *testing.T, srv *httptest.Server) *OrphanTokenChecker {
	t.Helper()
	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewOrphanTokenChecker(client, log.New(os.Stderr, "", 0))
	if err != nil {
		t.Fatalf("NewOrphanTokenChecker: %v", err)
	}
	return checker
}

func TestNewOrphanTokenChecker_NilClient(t *testing.T) {
	_, err := NewOrphanTokenChecker(nil, log.New(os.Stderr, "", 0))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewOrphanTokenChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://localhost", "tok")
	_, err := NewOrphanTokenChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestIsOrphanToken_Success_Orphan(t *testing.T) {
	srv := newMockOrphanServer(t, true)
	defer srv.Close()

	checker := newOrphanChecker(t, srv)
	info, err := checker.IsOrphanToken("some-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Orphan {
		t.Errorf("expected orphan=true, got false")
	}
	if info.TTL != 3600 {
		t.Errorf("expected TTL=3600, got %d", info.TTL)
	}
}

func TestIsOrphanToken_Success_NotOrphan(t *testing.T) {
	srv := newMockOrphanServer(t, false)
	defer srv.Close()

	checker := newOrphanChecker(t, srv)
	info, err := checker.IsOrphanToken("some-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Orphan {
		t.Errorf("expected orphan=false, got true")
	}
}

func TestIsOrphanToken_EmptyToken(t *testing.T) {
	srv := newMockOrphanServer(t, false)
	defer srv.Close()

	checker := newOrphanChecker(t, srv)
	_, err := checker.IsOrphanToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}
