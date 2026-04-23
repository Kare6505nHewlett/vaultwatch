package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockTokenMetadataServer(t *testing.T, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func newTokenMetadataChecker(t *testing.T, addr string) *TokenMetadataChecker {
	t.Helper()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewTokenMetadataChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewTokenMetadataChecker: %v", err)
	}
	return checker
}

func TestNewTokenMetadataChecker_NilClient(t *testing.T) {
	_, err := NewTokenMetadataChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewTokenMetadataChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://localhost:8200", "tok")
	_, err := NewTokenMetadataChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetTokenMetadata_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"display_name": "test-token",
			"policies":    []string{"default", "admin"},
			"entity_id":   "abc-123",
			"orphan":      false,
		},
	}
	server := newMockTokenMetadataServer(t, http.StatusOK, payload)
	defer server.Close()

	checker := newTokenMetadataChecker(t, server.URL)
	meta, err := checker.GetTokenMetadata("test-accessor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.DisplayName != "test-token" {
		t.Errorf("expected display_name 'test-token', got %q", meta.DisplayName)
	}
	if len(meta.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(meta.Policies))
	}
}

func TestGetTokenMetadata_EmptyAccessor(t *testing.T) {
	server := newMockTokenMetadataServer(t, http.StatusOK, nil)
	defer server.Close()

	checker := newTokenMetadataChecker(t, server.URL)
	_, err := checker.GetTokenMetadata("")
	if err == nil {
		t.Fatal("expected error for empty accessor")
	}
}

func TestGetTokenMetadata_NotFound(t *testing.T) {
	server := newMockTokenMetadataServer(t, http.StatusNotFound, nil)
	defer server.Close()

	checker := newTokenMetadataChecker(t, server.URL)
	_, err := checker.GetTokenMetadata("missing-accessor")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
