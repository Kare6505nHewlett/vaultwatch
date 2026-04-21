package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockStepDownServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(statusCode)
	}))
}

func newStepDownChecker(t *testing.T, addr string) *StepDownChecker {
	t.Helper()
	client, err := NewClient(addr, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := NewStepDownChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewStepDownChecker: %v", err)
	}
	return checker
}

func TestNewStepDownChecker_NilClient(t *testing.T) {
	_, err := NewStepDownChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewStepDownChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1:8200", "tok")
	_, err := NewStepDownChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestCheckStepDownEndpoint_Success(t *testing.T) {
	srv := newMockStepDownServer(http.StatusNoContent)
	defer srv.Close()

	checker := newStepDownChecker(t, srv.URL)
	result, err := checker.CheckStepDownEndpoint(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Reachable {
		t.Error("expected endpoint to be reachable")
	}
	if result.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", result.StatusCode)
	}
}

func TestCheckStepDownEndpoint_ServerError(t *testing.T) {
	srv := newMockStepDownServer(http.StatusInternalServerError)
	defer srv.Close()

	checker := newStepDownChecker(t, srv.URL)
	result, err := checker.CheckStepDownEndpoint(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", result.StatusCode)
	}
}
