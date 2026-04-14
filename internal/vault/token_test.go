package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newMockTokenServer(t *testing.T, statusCode int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestGetTokenInfo_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"accessor":  "abc123",
			"ttl":       3600,
			"renewable": true,
			"policies":  []string{"default", "admin"},
		},
	}
	server := newMockTokenServer(t, http.StatusOK, payload)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	info, err := client.GetTokenInfo(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if info.Accessor != "abc123" {
		t.Errorf("expected accessor 'abc123', got '%s'", info.Accessor)
	}
	if info.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", info.TTL)
	}
	if !info.Renewable {
		t.Error("expected token to be renewable")
	}
	if len(info.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(info.Policies))
	}
	if info.ExpireTime.Before(time.Now()) {
		t.Error("expected ExpireTime to be in the future")
	}
}

func TestGetTokenInfo_ServerError(t *testing.T) {
	server := newMockTokenServer(t, http.StatusForbidden, nil)
	defer server.Close()

	client, err := NewClient(server.URL, "bad-token")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	_, err = client.GetTokenInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for forbidden response, got nil")
	}
}
