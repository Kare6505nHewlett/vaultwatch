package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestSlackServer(t *testing.T, statusCode int, capturedBody *slackPayload) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capturedBody != nil {
			_ = json.NewDecoder(r.Body).Decode(capturedBody)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewSlackNotifier_EmptyURL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := NewSlackNotifier("", logger)
	if err == nil {
		t.Fatal("expected error for empty webhook URL")
	}
}

func TestNewSlackNotifier_NilLogger(t *testing.T) {
	_, err := NewSlackNotifier("http://example.com", nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestSlackNotifier_Send_Success(t *testing.T) {
	var captured slackPayload
	server := newTestSlackServer(t, http.StatusOK, &captured)
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	notifier, err := NewSlackNotifier(server.URL, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	alert := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/db/password",
		Message:    "expires soon",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	if err := notifier.Send(alert); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if captured.Text == "" {
		t.Error("expected non-empty slack message text")
	}
}

func TestSlackNotifier_Send_Non2xx(t *testing.T) {
	server := newTestSlackServer(t, http.StatusInternalServerError, nil)
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	notifier, err := NewSlackNotifier(server.URL, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	alert := Alert{
		Level:      LevelExpired,
		SecretPath: "secret/api/key",
		Message:    "already expired",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}

	if err := notifier.Send(alert); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}
