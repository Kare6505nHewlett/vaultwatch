package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockTokenRenewServer(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/token/renew-self" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"auth": map[string]interface{}{
					"client_token":   "s.renewed123",
					"lease_duration": 3600,
					"renewable":      true,
				},
			})
		}
	}))
}

func newTokenRenewer(t *testing.T, srv *httptest.Server) *TokenRenewer {
	t.Helper()
	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	r, err := NewTokenRenewer(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTokenRenewer: %v", err)
	}
	return r
}

func TestNewTokenRenewer_NilClient(t *testing.T) {
	_, err := NewTokenRenewer(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTokenRenewer_NilLogger(t *testing.T) {
	srv := newMockTokenRenewServer(t, http.StatusOK)
	defer srv.Close()
	client, _ := NewClient(srv.URL, "tok")
	_, err := NewTokenRenewer(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestRenewSelf_Success(t *testing.T) {
	srv := newMockTokenRenewServer(t, http.StatusOK)
	defer srv.Close()
	r := newTokenRenewer(t, srv)

	result, err := r.RenewSelf(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ClientToken != "s.renewed123" {
		t.Errorf("expected client_token s.renewed123, got %s", result.ClientToken)
	}
	if result.LeaseDuration != 3600 {
		t.Errorf("expected lease_duration 3600, got %d", result.LeaseDuration)
	}
	if !result.Renewable {
		t.Error("expected renewable true")
	}
}

func TestRenewSelf_ServerError(t *testing.T) {
	srv := newMockTokenRenewServer(t, http.StatusForbidden)
	defer srv.Close()
	r := newTokenRenewer(t, srv)

	_, err := r.RenewSelf(context.Background(), 3600)
	if err == nil {
		t.Fatal("expected error on server error")
	}
}
