package vault

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newMockMountServer(t *testing.T, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func newMountChecker(t *testing.T, addr string) *MountChecker {
	t.Helper()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewMountChecker(client, log.New(os.Stderr, "", 0))
	if err != nil {
		t.Fatalf("NewMountChecker: %v", err)
	}
	return checker
}

func TestNewMountChecker_NilClient(t *testing.T) {
	_, err := NewMountChecker(nil, log.New(os.Stderr, "", 0))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewMountChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewMountChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestListMounts_Success(t *testing.T) {
	payload := map[string]interface{}{
		"secret/": map[string]string{"type": "kv", "description": "KV store", "accessor": "abc123"},
		"sys/":    map[string]string{"type": "system", "description": "System", "accessor": "def456"},
	}
	srv := newMockMountServer(t, http.StatusOK, payload)
	defer srv.Close()

	checker := newMountChecker(t, srv.URL)
	mounts, err := checker.ListMounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mounts) != 2 {
		t.Errorf("expected 2 mounts, got %d", len(mounts))
	}
}

func TestListMounts_ServerError(t *testing.T) {
	srv := newMockMountServer(t, http.StatusForbidden, nil)
	defer srv.Close()

	checker := newMountChecker(t, srv.URL)
	_, err := checker.ListMounts(context.Background())
	if err == nil {
		t.Fatal("expected error on non-200 response")
	}
}
