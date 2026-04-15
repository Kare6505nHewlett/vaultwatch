package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockHealthServer(t *testing.T, status HealthStatus, code int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(status)
	}))
}

func TestNewHealthChecker_NilClient(t *testing.T) {
	_, err := NewHealthChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewHealthChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "token")
	_, err := NewHealthChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestHealthCheck_Success(t *testing.T) {
	expected := HealthStatus{
		Initialized: true,
		Sealed:      false,
		Standby:     false,
		Version:     "1.15.0",
		ClusterName: "vault-cluster",
	}
	server := newMockHealthServer(t, expected, http.StatusOK)
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	checker, _ := NewHealthChecker(client, zap.NewNop())

	status, err := checker.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Initialized {
		t.Error("expected initialized to be true")
	}
	if status.Sealed {
		t.Error("expected sealed to be false")
	}
	if status.Version != "1.15.0" {
		t.Errorf("expected version 1.15.0, got %s", status.Version)
	}
}

func TestHealthCheck_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	checker, _ := NewHealthChecker(client, zap.NewNop())

	_, err := checker.Check(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}
