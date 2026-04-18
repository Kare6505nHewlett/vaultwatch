package alert_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/monitor"
)

// fakeNotifier records calls and optionally returns an error.
type fakeNotifier struct {
	called int
	err    error
}

func (f *fakeNotifier) Send(_ context.Context, _ monitor.CheckResult) error {
	f.called++
	return f.err
}

func baseResult() monitor.CheckResult {
	return monitor.CheckResult{
		Path:      "secret/my-app/db",
		TTL:       72 * time.Hour,
		ExpiresAt: time.Now().Add(72 * time.Hour),
		Status:    monitor.StatusOK,
	}
}

func TestMultiNotifier_AllSucceed(t *testing.T) {
	a, b := &fakeNotifier{}, &fakeNotifier{}
	mn := alert.NewMultiNotifier(zap.NewNop(), a, b)

	err := mn.Send(context.Background(), baseResult())

	assert.NoError(t, err)
	assert.Equal(t, 1, a.called)
	assert.Equal(t, 1, b.called)
}

func TestMultiNotifier_OneFailsContinues(t *testing.T) {
	sentinel := errors.New("slack down")
	a := &fakeNotifier{err: sentinel}
	b := &fakeNotifier{}
	mn := alert.NewMultiNotifier(zap.NewNop(), a, b)

	err := mn.Send(context.Background(), baseResult())

	assert.ErrorIs(t, err, sentinel)
	// second notifier must still have been called
	assert.Equal(t, 1, b.called)
}

func TestMultiNotifier_AllFail(t *testing.T) {
	errA := errors.New("notifier A failed")
	errB := errors.New("notifier B failed")
	a := &fakeNotifier{err: errA}
	b := &fakeNotifier{err: errB}
	mn := alert.NewMultiNotifier(zap.NewNop(), a, b)

	err := mn.Send(context.Background(), baseResult())

	// both errors should be surfaced
	assert.ErrorIs(t, err, errA)
	assert.ErrorIs(t, err, errB)
	assert.Equal(t, 1, a.called)
	assert.Equal(t, 1, b.called)
}

func TestMultiNotifier_NilLogger(t *testing.T) {
	n := &fakeNotifier{}
	mn := alert.NewMultiNotifier(nil, n)

	assert.NotNil(t, mn)
	err := mn.Send(context.Background(), baseResult())
	assert.NoError(t, err)
}

func TestMultiNotifier_NoNotifiers(t *testing.T) {
	mn := alert.NewMultiNotifier(zap.NewNop())
	err := mn.Send(context.Background(), baseResult())
	assert.NoError(t, err)
}
