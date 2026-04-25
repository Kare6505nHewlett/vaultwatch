package vault

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newMockTokenRolesServer(t *testing.T, roleName string, statusCode int, role interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if role != nil {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": role})
		}
	}))
}

func newTokenRolesChecker(t *testing.T, addr string) *TokenRolesChecker {
	t.Helper()
	client, _ := NewClient(addr, "test-token")
	logger := log.New(os.Stdout, "", 0)
	checker, err := NewTokenRolesChecker(client, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return checker
}

func TestNewTokenRolesChecker_NilClient(t *testing.T) {
	_, err := NewTokenRolesChecker(nil, log.New(os.Stdout, "", 0))
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTokenRolesChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewTokenRolesChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetTokenRole_Success(t *testing.T) {
	role := map[string]interface{}{
		"allowed_policies": []string{"default"},
		"orphan":           true,
		"renewable":        true,
		"token_max_ttl":    3600,
		"explicit_max_ttl": 7200,
	}
	server := newMockTokenRolesServer(t, "myrole", http.StatusOK, role)
	defer server.Close()

	checker := newTokenRolesChecker(t, server.URL)
	result, err := checker.GetTokenRole("myrole")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "myrole" {
		t.Errorf("expected name 'myrole', got %q", result.Name)
	}
	if !result.Orphan {
		t.Error("expected orphan=true")
	}
	if result.MaxTTL != 3600 {
		t.Errorf("expected MaxTTL=3600, got %d", result.MaxTTL)
	}
}

func TestGetTokenRole_NotFound(t *testing.T) {
	server := newMockTokenRolesServer(t, "missing", http.StatusNotFound, nil)
	defer server.Close()

	checker := newTokenRolesChecker(t, server.URL)
	_, err := checker.GetTokenRole("missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestGetTokenRole_EmptyName(t *testing.T) {
	server := newMockTokenRolesServer(t, "", http.StatusOK, nil)
	defer server.Close()

	checker := newTokenRolesChecker(t, server.URL)
	_, err := checker.GetTokenRole("")
	if err == nil {
		t.Fatal("expected error for empty role name")
	}
}
