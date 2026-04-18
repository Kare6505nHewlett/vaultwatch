package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockWrappingServer(t *testing.T, status int, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func newWrappingChecker(t *testing.T, addr string) *WrappingChecker {
	t.Helper()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewWrappingChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewWrappingChecker: %v", err)
	}
	return checker
}

func TestNewWrappingChecker_NilClient(t *testing.T) {
	_, err := NewWrappingChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewWrappingChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewWrappingChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestLookupWrappingToken_EmptyToken(t *testing.T) {
	checker := newWrappingChecker(t, "http://127.0.0.1:8200")
	_, err := checker.LookupWrappingToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestLookupWrappingToken_Success(t *testing.T) {
	body := map[string]any{
		"data": map[string]any{
			"token":         "wrapping-token-abc",
			"accessor":      "acc-123",
			"ttl":           300,
			"creation_time": "2024-01-01T00:00:00Z",
			"creation_path": "secret/data/myapp",
		},
	}
	srv := newMockWrappingServer(t, http.StatusOK, body)
	defer srv.Close()

	checker := newWrappingChecker(t, srv.URL)
	info, err := checker.LookupWrappingToken("wrapping-token-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Accessor != "acc-123" {
		t.Errorf("expected accessor acc-123, got %s", info.Accessor)
	}
	if info.CreationPath != "secret/data/myapp" {
		t.Errorf("unexpected creation path: %s", info.CreationPath)
	}
}

func TestLookupWrappingToken_ServerError(t *testing.T) {
	srv := newMockWrappingServer(t, http.StatusForbidden, nil)
	defer srv.Close()

	checker := newWrappingChecker(t, srv.URL)
	_, err := checker.LookupWrappingToken("bad-token")
	if err == nil {
		t.Fatal("expected error on non-200 status")
	}
}
