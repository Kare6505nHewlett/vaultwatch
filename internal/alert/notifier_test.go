package alert

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"
)

func newTestNotifier(buf *bytes.Buffer) *LogNotifier {
	logger := log.New(buf, "", 0)
	return NewLogNotifier(logger)
}

func TestSend_WarningAlert(t *testing.T) {
	var buf bytes.Buffer
	n := newTestNotifier(&buf)

	expiry := time.Now().Add(30 * time.Minute)
	a := Alert{
		SecretPath: "secret/my-app/db",
		Level:      LevelWarning,
		ExpiresAt:  expiry,
		TimeLeft:   30 * time.Minute,
	}

	err := n.Send(a)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[WARNING]") {
		t.Errorf("expected WARNING in output, got: %s", out)
	}
	if !strings.Contains(out, "secret/my-app/db") {
		t.Errorf("expected secret path in output, got: %s", out)
	}
	if !strings.Contains(out, "30m0s") {
		t.Errorf("expected time left in output, got: %s", out)
	}
}

func TestSend_ExpiredAlert(t *testing.T) {
	var buf bytes.Buffer
	n := newTestNotifier(&buf)

	expiry := time.Now().Add(-5 * time.Minute)
	a := Alert{
		SecretPath: "secret/my-app/api-key",
		Level:      LevelExpired,
		ExpiresAt:  expiry,
		TimeLeft:   0
	}

	err := n.Send(a)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[EXPIRED]") {
		t.Errorf("expected EXPIRED in output, got: %s", out)
	}
	if !strings.Contains(out, "secret/my-app/api-key") {
		t.Errorf("expected secret path in output, got: %s", out)
	}
}

func TestNewLogNotifier_NilLogger(t *testing.T) {
	n := NewLogNotifier(nil)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
	if n.logger == nil {
		t.Fatal("expected non-nil internal logger")
	}
}
