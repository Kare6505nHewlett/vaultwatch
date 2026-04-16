package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newMockSealServer(t *testing.T, status SealStatus, code int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/seal-status" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(code)
		if code == http.StatusOK {
			_ = json.NewEncoder(w).Encode(status)
		}
	}))
}

func TestNewSealChecker_NilClient(t *testing.T) {
	_, err := NewSealChecker(nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewSealChecker_NilLogger(t *testing.T) {
	client, _ := NewClient("http://localhost", "token")
	_, err := NewSealChecker(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestGetSealStatus_Success(t *testing.T) {
	expected := SealStatus{Sealed: false, Initialized: true, Version: "1.15.0"}
	srv := newMockSealServer(t, expected, http.StatusOK)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token")
	checker, _ := NewSealChecker(client, zap.NewNop())

	status, err := checker.GetSealStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Sealed != expected.Sealed {
		t.Errorf("expected sealed=%v, got %v", expected.Sealed, status.Sealed)
	}
	if status.Version != expected.Version {
		t.Errorf("expected version=%s, got %s", expected.Version, status.Version)
	}
}

func TestGetSealStatus_ServerError(t *testing.T) {
	srv := newMockSealServer(t, SealStatus{}, http.StatusInternalServerError)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token")
	checker, _ := NewSealChecker(client, zap.NewNop())

	_, err := checker.GetSealStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}
