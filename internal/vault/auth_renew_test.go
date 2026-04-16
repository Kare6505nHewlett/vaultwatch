package vault

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newMockAuthRenewServer(statusCode int, ttl int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"auth": map[string]any{"lease_duration": ttl},
			})
		}
	}))
}

func newAuthRenewer(t *testing.T, addr string) *AuthRenewer {
	t.Helper()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	r, err := NewAuthRenewer(client, logger)
	if err != nil {
		t.Fatalf("NewAuthRenewer: %v", err)
	}
	return r
}

func TestNewAuthRenewer_NilClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err := NewAuthRenewer(nil, logger)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewAuthRenewer_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewAuthRenewer(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestRenewSelf_Success(t *testing.T) {
	srv := newMockAuthRenewServer(http.StatusOK, 3600)
	defer srv.Close()

	r := newAuthRenewer(t, srv.URL)
	ttl, err := r.RenewSelf(context.Background(), 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ttl != 3600 {
		t.Errorf("expected ttl 3600, got %d", ttl)
	}
}

func TestRenewSelf_ServerError(t *testing.T) {
	srv := newMockAuthRenewServer(http.StatusForbidden, 0)
	defer srv.Close()

	r := newAuthRenewer(t, srv.URL)
	_, err := r.RenewSelf(context.Background(), 3600)
	if err == nil {
		t.Fatal("expected error on non-200 response")
	}
}
