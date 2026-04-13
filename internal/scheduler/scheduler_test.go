package scheduler_test

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/scheduler"
)

func newTestLogger(buf *bytes.Buffer) *log.Logger {
	return log.New(buf, "", 0)
}

func TestNew_ReturnsScheduler(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	s := scheduler.New(nil, nil, []scheduler.Job{}, logger)
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestRun_CancelStopsScheduler(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	s := scheduler.New(nil, nil, []scheduler.Job{}, logger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		s.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop after context cancellation")
	}
}

func TestRun_AlreadyCancelledContext(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	s := scheduler.New(nil, nil, []scheduler.Job{}, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Run is called

	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// success — scheduler should exit immediately
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop with pre-cancelled context")
	}
}

func TestJob_Fields(t *testing.T) {
	job := scheduler.Job{
		SecretPath: "secret/myapp/db",
		Interval:   30 * time.Second,
	}

	if job.SecretPath != "secret/myapp/db" {
		t.Errorf("unexpected SecretPath: %s", job.SecretPath)
	}
	if job.Interval != 30*time.Second {
		t.Errorf("unexpected Interval: %s", job.Interval)
	}
}
