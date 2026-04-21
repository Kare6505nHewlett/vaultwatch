package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockCapabilitiesServer(t *testing.T, capabilities []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/capabilities-self" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"capabilities": capabilities,
			},
		})
	}))
}

func newCapabilitiesChecker(t *testing.T, serverURL string) *CapabilitiesChecker {
	t.Helper()
	client, err := NewClient(serverURL, "test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	checker, err := NewCapabilitiesChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("failed to create capabilities checker: %v", err)
	}
	return checker
}

func TestNewCapabilitiesChecker_NilClient(t *testing.T) {
	_, err := NewCapabilitiesChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewCapabilitiesChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://localhost:8200", "token")
	_, err := NewCapabilitiesChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestCheckCapabilities_Success(t *testing.T) {
	expected := []string{"read", "list"}
	server := newMockCapabilitiesServer(t, expected)
	defer server.Close()

	checker := newCapabilitiesChecker(t, server.URL)
	result, err := checker.CheckCapabilities(context.Background(), "secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Capabilities) != len(expected) {
		t.Errorf("expected %d capabilities, got %d", len(expected), len(result.Capabilities))
	}
	if result.Path != "secret/data/myapp" {
		t.Errorf("expected path %q, got %q", "secret/data/myapp", result.Path)
	}
}

func TestCheckCapabilities_EmptyPath(t *testing.T) {
	server := newMockCapabilitiesServer(t, []string{"read"})
	defer server.Close()

	checker := newCapabilitiesChecker(t, server.URL)
	_, err := checker.CheckCapabilities(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}
