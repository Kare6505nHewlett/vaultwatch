package scheduler_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/scheduler"
)

// mockHealthMonitor implements a simple health monitor for testing.
type mockHealthMonitor struct {
	callCount int
	shouldFail bool
	result     monitor.HealthResult
}

func (m *mockHealthMonitor) Check(ctx context.Context) (monitor.HealthResult, error) {
	m.callCount++
	if m.shouldFail {
		return monitor.HealthResult{}, assert.AnError
	}
	return m.result, nil
}

func newHealthJobLogger()	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	 slog.LevelDebug,
	}))
}

func TestNewHealthJob_NilMonitor(t *testing.T) {
	logger := newHealthJobLogger()
	job, err := scheduler.NewHealthJob(nil, logger)
	require.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "monitor")
}

func TestNewHealthJob_NilLogger(t *testing.T) {
	mon := &mockHealthMonitor{}
	job, err := scheduler.NewHealthJob(mon, nil)
	require.Error(t, err)
	assert.Nil(t, job)
	assert.Contains(t, err.Error(), "logger")
}

func TestNewHealthJob_Valid(t *testing.T) {
	mon := &mockHealthMonitor{}
	logger := newHealthJobLogger()

	job, err := scheduler.NewHealthJob(mon, logger)
	require.NoError(t, err)
	assert.NotNil(t, job)
}

func TestHealthJob_RunsMonitorCheck(t *testing.T) {
	mon := &mockHealthMonitor{
		result: monitor.HealthResult{
			Healthy:     true,
			Initialized: true,
			Sealed:      false,
		},
	}
	logger := newHealthJobLogger()

	job, err := scheduler.NewHealthJob(mon, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sched := scheduler.New(logger)
	sched.Register(scheduler.Job{
		Name:     "health-check",
		Interval: 50 * time.Millisecond,
		Run:      job.Run,
	})

	go func() {
		_ = sched.Run(ctx)
	}()

	// Allow at least one execution.
	time.Sleep(150 * time.Millisecond)
	cancel()

	assert.GreaterOrEqual(t, mon.callCount, 1)
}

func TestHealthJob_HandlesMonitorError(t *testing.T) {
	mon := &mockHealthMonitor{
		shouldFail: true,
	}
	logger := newHealthJobLogger()

	job, err := scheduler.NewHealthJob(mon, logger)
	require.NoError(t, err)

	ctx := context.Background()
	// Should not panic even when monitor returns an error.
	assert.NotPanics(t, func() {
		job.Run(ctx)
	})
	assert.Equal(t, 1, mon.callCount)
}
