package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockSnapshotServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func newSnapshotChecker(t *testing.T, addr string) *SnapshotChecker {
	t.Helper()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewSnapshotChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewSnapshotChecker: %v", err)
	}
	return checker
}

func TestNewSnapshotChecker_NilClient(t *testing.T) {
	_, err := NewSnapshotChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewSnapshotChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewSnapshotChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestCheckSnapshot_Success(t *testing.T) {
	srv := newMockSnapshotServer(http.StatusOK, "snapshotdata")
	defer srv.Close()

	checker := newSnapshotChecker(t, srv.URL)
	result, err := checker.CheckSnapshot(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Available {
		t.Error("expected snapshot to be available")
	}
	if result.Bytes == 0 {
		t.Error("expected non-zero bytes")
	}
}

func TestCheckSnapshot_ServerError(t *testing.T) {
	srv := newMockSnapshotServer(http.StatusInternalServerError, "")
	defer srv.Close()

	checker := newSnapshotChecker(t, srv.URL)
	result, err := checker.CheckSnapshot(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Available {
		t.Error("expected snapshot to be unavailable on 500")
	}
	if result.Error == "" {
		t.Error("expected error message")
	}
}
