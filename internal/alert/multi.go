package alert

import (
	"context"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// Notifier defines the interface for sending alerts.
type Notifier interface {
	Send(ctx context.Context, result monitor.CheckResult) error
}

// MultiNotifier fans out alerts to multiple Notifier implementations.
type MultiNotifier struct {
	notifiers []Notifier
	logger    *zap.Logger
}

// NewMultiNotifier creates a MultiNotifier that dispatches to all provided notifiers.
func NewMultiNotifier(logger *zap.Logger, notifiers ...Notifier) *MultiNotifier {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &MultiNotifier{
		notifiers: notifiers,
		logger:    logger,
	}
}

// Send dispatches the alert to every registered notifier.
// Errors are logged but do not stop delivery to remaining notifiers.
func (m *MultiNotifier) Send(ctx context.Context, result monitor.CheckResult) error {
	var firstErr error
	for _, n := range m.notifiers {
		if err := n.Send(ctx, result); err != nil {
			m.logger.Error("notifier failed to send alert",
				zap.Error(err),
				zap.String("path", result.Path),
			)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
