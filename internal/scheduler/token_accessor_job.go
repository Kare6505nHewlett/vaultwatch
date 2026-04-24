package scheduler

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"go.uber.org/zap"
)

// TokenAccessorJob is a scheduler job that checks token accessors.
type TokenAccessorJob struct {
	monitor *monitor.TokenAccessorMonitor
	logger  *zap.Logger
}

// NewTokenAccessorJob creates a new TokenAccessorJob.
func NewTokenAccessorJob(m *monitor.TokenAccessorMonitor, logger *zap.Logger) (*TokenAccessorJob, error) {
	if m == nil {
		return nil, fmt.Errorf("token accessor monitor must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenAccessorJob{monitor: m, logger: logger}, nil
}

// Name returns the job identifier.
func (j *TokenAccessorJob) Name() string {
	return "token-accessor-check"
}

// Run executes the token accessor check and logs results.
func (j *TokenAccessorJob) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	results := j.monitor.Check()
	for _, r := range results {
		switch r.Status {
		case "expired":
			j.logger.Error("token accessor expired",
				zap.String("accessor", r.Accessor),
				zap.String("display_name", r.DisplayName),
				zap.String("message", r.Message),
			)
		case "warning":
			j.logger.Warn("token accessor expiring soon",
				zap.String("accessor", r.Accessor),
				zap.String("display_name", r.DisplayName),
				zap.String("message", r.Message),
			)
		case "error":
			j.logger.Error("token accessor check failed",
				zap.String("accessor", r.Accessor),
				zap.String("message", r.Message),
			)
		default:
			j.logger.Info("token accessor ok",
				zap.String("accessor", r.Accessor),
				zap.String("display_name", r.DisplayName),
				zap.Duration("ttl", r.TTL),
			)
		}
	}
	return nil
}
