package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockReplicationServer(status int, body interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(body)
	}))
}

func newReplicationChecker(t *testing.T, addr string) *ReplicationChecker {
	t.Helper()
	client, _ := NewClient(addr, "test-token")
	logger := zap.NewNop()
	rc, err := NewReplicationChecker(client, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return rc
}

func TestNewReplicationChecker_NilClient(t *testing.T) {
	_, err := NewReplicationChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewReplicationChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewReplicationChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetReplicationStatus_Success(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"dr":          map[string]string{"mode": "primary", "state": "running"},
			"performance": map[string]string{"mode": "disabled", "state": ""},
		},
	}
	server := newMockReplicationServer(http.StatusOK, payload)
	defer server.Close()

	rc := newReplicationChecker(t, server.URL)
	result, err := rc.GetReplicationStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data.DR.Mode != "primary" {
		t.Errorf("expected dr mode 'primary', got %q", result.Data.DR.Mode)
	}
}

func TestGetReplicationStatus_ServerError(t *testing.T) {
	server := newMockReplicationServer(http.StatusInternalServerError, nil)
	defer server.Close()

	rc := newReplicationChecker(t, server.URL)
	_, err := rc.GetReplicationStatus()
	if err == nil {
		t.Fatal("expected error on server error")
	}
}
