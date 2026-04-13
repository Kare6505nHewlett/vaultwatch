package alert_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/monitor"
)

func newTestWebhookServer(t *testing.T, statusCode int, received *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if received != nil {
			_ = json.NewDecoder(r.Body).Decode(received)
		}
		w.WriteHeader(statusCode)
	}))
}

func newWebhookLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func TestNewWebhookNotifier_EmptyURL(t *testing.T) {
	_, err := alert.NewWebhookNotifier("", newWebhookLogger())
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewWebhookNotifier_NilLogger(t *testing.T) {
	_, err := alert.NewWebhookNotifier("http://example.com", nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received map[string]interface{}
	server := newTestWebhookServer(t, http.StatusOK, &received)
	defer server.Close()

	n, err := alert.NewWebhookNotifier(server.URL, newWebhookLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := monitor.CheckResult{
		SecretPath: "secret/my-app/db",
		Level:      monitor.LevelWarning,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		Message:    "expires soon",
	}

	if err := n.Send(result); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	if received["secret"] != "secret/my-app/db" {
		t.Errorf("expected secret path in payload, got %v", received["secret"])
	}
	if received["level"] != string(monitor.LevelWarning) {
		t.Errorf("expected level warning in payload, got %v", received["level"])
	}
}

func TestWebhookNotifier_Send_Non2xx(t *testing.T) {
	server := newTestWebhookServer(t, http.StatusInternalServerError, nil)
	defer server.Close()

	n, err := alert.NewWebhookNotifier(server.URL, newWebhookLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := monitor.CheckResult{
		SecretPath: "secret/api",
		Level:      monitor.LevelExpired,
		ExpiresAt:  time.Now().Add(-time.Hour),
		Message:    "already expired",
	}

	if err := n.Send(result); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}
