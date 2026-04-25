package vault_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

func newMockTokenTTLServer(t *testing.T, statusCode int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func newTokenTTLChecker(t *testing.T, addr string) *vault.TokenTTLChecker {
	t.Helper()
	client, err := vault.NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewTokenTTLChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTokenTTLChecker: %v", err)
	}
	return checker
}

func TestNewTokenTTLChecker_NilClient(t *testing.T) {
	_, err := vault.NewTokenTTLChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTokenTTLChecker_NilLogger(t *testing.T) {
	client, _ := vault.NewClient("http://127.0.0.1:8200", "tok")
	_, err := vault.NewTokenTTLChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetTokenTTL_Success(t *testing.T) {
	expireTime := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"ttl":          7200,
			"creation_ttl": 86400,
			"expire_time":  expireTime,
			"renewable":    true,
		},
	}
	srv := newMockTokenTTLServer(t, http.StatusOK, payload)
	defer srv.Close()

	checker := newTokenTTLChecker(t, srv.URL)
	info, err := checker.GetTokenTTL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TTL != 7200*time.Second {
		t.Errorf("expected TTL 7200s, got %v", info.TTL)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
	if info.ExpireTime == nil {
		t.Error("expected expire_time to be set")
	}
}

func TestGetTokenTTL_ServerError(t *testing.T) {
	srv := newMockTokenTTLServer(t, http.StatusForbidden, nil)
	defer srv.Close()

	checker := newTokenTTLChecker(t, srv.URL)
	_, err := checker.GetTokenTTL()
	if err == nil {
		t.Fatal("expected error on non-200 status")
	}
}
