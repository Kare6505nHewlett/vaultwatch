package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newMockVaultServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func TestNewClient_InvalidAddress(t *testing.T) {
	// Vault API client creation itself rarely fails on bad address,
	// but we ensure no panic occurs.
	c, err := NewClient("http://127.0.0.1:1", "test-token")
	if err != nil {
		t.Fatalf("expected no error creating client, got: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGetSecretLease_Success(t *testing.T) {
	server := newMockVaultServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/secret/myapp/db" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"lease_id":       "secret/myapp/db/abc123",
			"lease_duration": 3600,
			"renewable":      true,
			"data": map[string]string{
				"username": "admin",
				"password": "s3cr3t",
			},
		})
	})
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.GetSecretLease("secret/myapp/db")
	if err != nil {
		t.Fatalf("GetSecretLease: %v", err)
	}

	if info.Path != "secret/myapp/db" {
		t.Errorf("expected path %q, got %q", "secret/myapp/db", info.Path)
	}
	if info.LeaseTTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", info.LeaseTTL)
	}
	if info.IsToken {
		t.Error("expected IsToken=false for secret lease")
	}
	if info.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestGetSecretLease_NotFound(t *testing.T) {
	server := newMockVaultServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{}})
	})
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.GetSecretLease("secret/does/not/exist")
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}
