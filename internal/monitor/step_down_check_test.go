package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockStepDownCheckServer(code int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}))
}

func newStepDownMonitor(t *testing.T, addr string) *StepDownMonitor {
	t.Helper()
	client, err := vault.NewClient(addr, "tok")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	checker, err := vault.NewStepDownChecker(client, zap.NewNop())
	if err != nil {
		t.Fatalf("NewStepDownChecker: %v", err)
	}
	mon, err := NewStepDownMonitor(checker, zap.NewNop())
	if err != nil {
		t.Fatalf("NewStepDownMonitor: %v", err)
	}
	return mon
}

func TestNewStepDownMonitor_NilChecker(t *testing.T) {
	_, err := NewStepDownMonitor(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil checker")
	}
}

func TestNewStepDownMonitor_NilLogger(t *testing.T) {
	client, _ := vault.NewClient("http://127.0.0.1:8200", "tok")
	checker, _ := vault.NewStepDownChecker(client, zap.NewNop())
	_, err := NewStepDownMonitor(checker, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestStepDownMonitor_Healthy(t *testing.T) {
	srv := newMockStepDownCheckServer(http.StatusNoContent)
	defer srv.Close()

	mon := newStepDownMonitor(t, srv.URL)
	status, err := mon.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Healthy {
		t.Errorf("expected healthy, got message: %s", status.Message)
	}
}

func TestStepDownMonitor_Forbidden(t *testing.T) {
	srv := newMockStepDownCheckServer(http.StatusForbidden)
	defer srv.Close()

	mon := newStepDownMonitor(t, srv.URL)
	status, err := mon.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 403 is still considered healthy (endpoint exists, token lacks perms)
	if !status.Healthy {
		t.Errorf("expected healthy for 403, got: %s", status.Message)
	}
}

func TestStepDownMonitor_Unhealthy(t *testing.T) {
	srv := newMockStepDownCheckServer(http.StatusInternalServerError)
	defer srv.Close()

	mon := newStepDownMonitor(t, srv.URL)
	status, err := mon.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Healthy {
		t.Error("expected unhealthy for 500 response")
	}
}
