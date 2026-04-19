package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockLoginServer(t *testing.T, statusCode int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func newLoginChecker(t *testing.T, addr string) *LoginChecker {
	t.Helper()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	logger, _ := zap.NewDevelopment()
	checker, err := NewLoginChecker(client, logger)
	if err != nil {
		t.Fatalf("NewLoginChecker: %v", err)
	}
	return checker
}

func TestNewLoginChecker_NilClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := NewLoginChecker(nil, logger)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewLoginChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewLoginChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestLoginWithAppRole_Success(t *testing.T) {
	body := map[string]interface{}{
		"auth": map[string]interface{}{
			"client_token":   "s.newtoken",
			"accessor":       "acc123",
			"lease_duration": 3600,
			"renewable":      true,
		},
	}
	srv := newMockLoginServer(t, http.StatusOK, body)
	defer srv.Close()

	checker := newLoginChecker(t, srv.URL)
	result, err := checker.LoginWithAppRole("role-id", "secret-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ClientToken != "s.newtoken" {
		t.Errorf("expected token s.newtoken, got %s", result.ClientToken)
	}
	if !result.Renewable {
		t.Error("expected renewable to be true")
	}
}

func TestLoginWithAppRole_EmptyCredentials(t *testing.T) {
	srv := newMockLoginServer(t, http.StatusOK, nil)
	defer srv.Close()
	checker := newLoginChecker(t, srv.URL)
	_, err := checker.LoginWithAppRole("", "secret")
	if err == nil {
		t.Fatal("expected error for empty roleID")
	}
}

func TestLoginWithAppRole_ServerError(t *testing.T) {
	srv := newMockLoginServer(t, http.StatusForbidden, nil)
	defer srv.Close()
	checker := newLoginChecker(t, srv.URL)
	_, err := checker.LoginWithAppRole("role", "secret")
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}
