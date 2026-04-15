package vault_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockRotateServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func newRotateLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func TestNewRotator_NilClient(t *testing.T) {
	_, err := vault.NewRotator(nil, newRotateLogger())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

func TestNewRotator_NilLogger(t *testing.T) {
	server := newMockRotateServer(http.StatusOK)
	defer server.Close()

	client, _ := vault.NewClient(server.URL, "test-token")
	_, err := vault.NewRotator(client, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestRotateSecret_Success(t *testing.T) {
	server := newMockRotateServer(http.StatusNoContent)
	defer server.Close()

	client, _ := vault.NewClient(server.URL, "test-token")
	rotator, _ := vault.NewRotator(client, newRotateLogger())

	result, err := rotator.RotateSecret(context.Background(), "database")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got: %s", result.Message)
	}
	if result.Path != "database" {
		t.Errorf("expected path 'database', got %s", result.Path)
	}
}

func TestRotateSecret_ServerError(t *testing.T) {
	server := newMockRotateServer(http.StatusInternalServerError)
	defer server.Close()

	client, _ := vault.NewClient(server.URL, "test-token")
	rotator, _ := vault.NewRotator(client, newRotateLogger())

	result, err := rotator.RotateSecret(context.Background(), "database")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if result.Success {
		t.Error("expected failure result")
	}
}

func TestRotateSecret_EmptyPath(t *testing.T) {
	server := newMockRotateServer(http.StatusOK)
	defer server.Close()

	client, _ := vault.NewClient(server.URL, "test-token")
	rotator, _ := vault.NewRotator(client, newRotateLogger())

	_, err := rotator.RotateSecret(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty mount path")
	}
}
