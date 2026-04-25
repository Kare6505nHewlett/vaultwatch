package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockRevokeServer(t *testing.T, selfStatus, accessorStatus int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/token/revoke-self":
			w.WriteHeader(selfStatus)
		case "/v1/auth/token/revoke-accessor":
			w.WriteHeader(accessorStatus)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newRevoker(t *testing.T, addr string) *TokenRevoker {
	t.Helper()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	r, err := NewTokenRevoker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTokenRevoker: %v", err)
	}
	return r
}

func TestNewTokenRevoker_NilClient(t *testing.T) {
	_, err := NewTokenRevoker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTokenRevoker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewTokenRevoker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestRevokeSelf_Success(t *testing.T) {
	srv := newMockRevokeServer(t, http.StatusNoContent, http.StatusNoContent)
	defer srv.Close()

	r := newRevoker(t, srv.URL)
	if err := r.RevokeSelf(context.Background()); err != nil {
		t.Fatalf("RevokeSelf: %v", err)
	}
}

func TestRevokeSelf_ServerError(t *testing.T) {
	srv := newMockRevokeServer(t, http.StatusInternalServerError, http.StatusNoContent)
	defer srv.Close()

	r := newRevoker(t, srv.URL)
	if err := r.RevokeSelf(context.Background()); err == nil {
		t.Fatal("expected error on server error")
	}
}

func TestRevokeAccessor_Success(t *testing.T) {
	srv := newMockRevokeServer(t, http.StatusNoContent, http.StatusNoContent)
	defer srv.Close()

	r := newRevoker(t, srv.URL)
	if err := r.RevokeAccessor(context.Background(), "abc123"); err != nil {
		t.Fatalf("RevokeAccessor: %v", err)
	}
}

func TestRevokeAccessor_EmptyAccessor(t *testing.T) {
	srv := newMockRevokeServer(t, http.StatusNoContent, http.StatusNoContent)
	defer srv.Close()

	r := newRevoker(t, srv.URL)
	if err := r.RevokeAccessor(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty accessor")
	}
}

func TestRevokeAccessor_ServerError(t *testing.T) {
	srv := newMockRevokeServer(t, http.StatusNoContent, http.StatusForbidden)
	defer srv.Close()

	r := newRevoker(t, srv.URL)
	if err := r.RevokeAccessor(context.Background(), "abc123"); err == nil {
		t.Fatal("expected error on server error")
	}
}
