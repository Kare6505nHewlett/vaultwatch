package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newMockRenewServer(statusCode int, leaseDuration int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			body := map[string]interface{}{
				"lease_id":       "secret/data/myapp/db#abc123",
				"lease_duration": leaseDuration,
				"renewable":      true,
			}
			json.NewEncoder(w).Encode(body) //nolint:errcheck
		}
	}))
}

func TestNewRenewer_NilClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := NewRenewer(nil, logger)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewRenewer_NilLogger(t *testing.T) {
	server := newMockRenewServer(http.StatusOK, 3600)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = NewRenewer(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestRenewLease_Success(t *testing.T) {
	server := newMockRenewServer(http.StatusOK, 3600)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	renewer, err := NewRenewer(client, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := renewer.RenewLease(context.Background(), "secret/data/myapp/db", 1*time.Hour)
	if !result.Renewed {
		t.Errorf("expected Renewed=true, got false; err=%v", result.Error)
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if result.NewExpiry.IsZero() {
		t.Error("expected non-zero NewExpiry")
	}
}

func TestRenewLease_ServerError(t *testing.T) {
	server := newMockRenewServer(http.StatusInternalServerError, 0)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	renewer, _ := NewRenewer(client, logger)

	result := renewer.RenewLease(context.Background(), "secret/data/myapp/db", 1*time.Hour)
	if result.Renewed {
		t.Error("expected Renewed=false on server error")
	}
	if result.Error == nil {
		t.Error("expected non-nil error on server error")
	}
}
