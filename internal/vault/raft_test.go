package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockRaftServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func newRaftChecker(t *testing.T, serverURL string) *RaftChecker {
	t.Helper()
	client, err := NewClient(serverURL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewRaftChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewRaftChecker: %v", err)
	}
	return checker
}

func TestNewRaftChecker_NilClient(t *testing.T) {
	_, err := NewRaftChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewRaftChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://localhost:8200", "tok")
	_, err := NewRaftChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetRaftStatus_Success(t *testing.T) {
	body := map[string]interface{}{
		"data": map[string]interface{}{
			"leader": "node-1",
			"apply_index": 42,
			"commit_index": 42,
			"servers": []map[string]interface{}{
				{"node_id": "node-1", "address": "127.0.0.1:8201", "leader": true, "voter": true},
			},
		},
	}
	srv := newMockRaftServer(t, http.StatusOK, body)
	defer srv.Close()

	checker := newRaftChecker(t, srv.URL)
	status, err := checker.GetRaftStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Leader != "node-1" {
		t.Errorf("expected leader node-1, got %s", status.Leader)
	}
	if len(status.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(status.Servers))
	}
}

func TestGetRaftStatus_ServerError(t *testing.T) {
	srv := newMockRaftServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	checker := newRaftChecker(t, srv.URL)
	_, err := checker.GetRaftStatus()
	if err == nil {
		t.Fatal("expected error on server error")
	}
}
