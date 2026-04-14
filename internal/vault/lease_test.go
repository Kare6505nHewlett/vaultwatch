package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockLeaseServer(t *testing.T, leaseID string, ttl int, renewable bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/"+leaseID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"lease_id":       leaseID,
				"lease_duration": ttl,
				"renewable":      renewable,
				"data":           map[string]interface{}{},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestNewLeaseManager_NilClient(t *testing.T) {
	logger := zap.NewNop()
	_, err := NewLeaseManager(nil, logger)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewLeaseManager_NilLogger(t *testing.T) {
	server := newMockLeaseServer(t, "secret/data/test", 3600, true)
	defer server.Close()
	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = NewLeaseManager(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetLeaseInfo_EmptyLeaseID(t *testing.T) {
	server := newMockLeaseServer(t, "secret/data/test", 3600, true)
	defer server.Close()
	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	lm, err := NewLeaseManager(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewLeaseManager: %v", err)
	}
	_, err = lm.GetLeaseInfo(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty lease ID")
	}
}

func TestGetLeaseInfo_Success(t *testing.T) {
	path := "secret/data/myapp"
	server := newMockLeaseServer(t, path, 7200, true)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	lm, err := NewLeaseManager(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewLeaseManager: %v", err)
	}

	info, err := lm.GetLeaseInfo(context.Background(), path)
	if err != nil {
		t.Fatalf("GetLeaseInfo: %v", err)
	}
	if info.LeaseID != path {
		t.Errorf("expected lease ID %q, got %q", path, info.LeaseID)
	}
	if info.TTL.Seconds() != 7200 {
		t.Errorf("expected TTL 7200s, got %v", info.TTL)
	}
	if !info.Renewable {
		t.Error("expected lease to be renewable")
	}
	if info.ExpiresAt.IsZero() {
		t.Error("expected non-zero ExpiresAt")
	}
}
