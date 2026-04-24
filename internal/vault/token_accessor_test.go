package vault_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap/zaptest"
)

func newMockTokenAccessorServer(t *testing.T, statusCode int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func newTokenAccessorChecker(t *testing.T, serverURL string) *vault.TokenAccessorChecker {
	t.Helper()
	logger := zaptest.NewLogger(t)
	client, err := vault.NewClient(serverURL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewTokenAccessorChecker(client, logger)
	if err != nil {
		t.Fatalf("NewTokenAccessorChecker: %v", err)
	}
	return checker
}

func TestNewTokenAccessorChecker_NilClient(t *testing.T) {
	logger := zaptest.NewLogger(t)
	_, err := vault.NewTokenAccessorChecker(nil, logger)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTokenAccessorChecker_NilLogger(t *testing.T) {
	server := newMockTokenAccessorServer(t, 200, nil)
	defer server.Close()
	client, _ := vault.NewClient(server.URL, "tok")
	_, err := vault.NewTokenAccessorChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestLookupTokenByAccessor_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"accessor":         "abc123",
			"display_name":     "test-token",
			"policies":         []string{"default", "admin"},
			"ttl":              3600,
			"expire_time":      "2099-01-01T00:00:00Z",
		},
	}
	server := newMockTokenAccessorServer(t, 200, payload)
	defer server.Close()
	checker := newTokenAccessorChecker(t, server.URL)
	info, err := checker.LookupByAccessor("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", info.Accessor)
	}
	if info.DisplayName != "test-token" {
		t.Errorf("expected display_name test-token, got %s", info.DisplayName)
	}
}

func TestLookupTokenByAccessor_EmptyAccessor(t *testing.T) {
	server := newMockTokenAccessorServer(t, 200, nil)
	defer server.Close()
	checker := newTokenAccessorChecker(t, server.URL)
	_, err := checker.LookupByAccessor("")
	if err == nil {
		t.Fatal("expected error for empty accessor")
	}
}

func TestLookupTokenByAccessor_ServerError(t *testing.T) {
	server := newMockTokenAccessorServer(t, 500, nil)
	defer server.Close()
	checker := newTokenAccessorChecker(t, server.URL)
	_, err := checker.LookupByAccessor("abc123")
	if err == nil {
		t.Fatal("expected error on server 500")
	}
}
