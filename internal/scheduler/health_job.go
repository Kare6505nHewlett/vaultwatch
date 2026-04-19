package scheduler

import (
	"context"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"go.uber.org/zap"
)

// HealthJob is a scheduler Job that runs a Vault health check.
type HealthJob struct {
	monitor  *monitor.HealthMonitor
	logger   *zap.Logger
	onSealed func(result monitor.HealthResult)
}

// NewHealthJob creates a Job that periodically checks Vault health.
// onSealed is called (if non-nil) when the vault is found to be sealed or unhealthy.
func NewHealthJob(mon *monitor.HealthMonitor, interval time.Duration, logger *zap.Logger, onSealed func(monitor.HealthResult)) Job {
	hj := &HealthJob{
		monitor:  mon,
		logger:   logger,
		onSealed: onSealed,
	}
	return Job{
		Name:     "vault-health",
		Interval: interval,
		Run: func(ctx context.Context) error {
			return hj.run(ctx)
		},
	}
}

func (h *HealthJob) run(ctx context.Context) error {
	result := h.monitor.Run(ctx)
	if result.Err != nil {
		h.logger.Error("health job encountered error", zap.Error(result.Err))
		h.notify(result)
		return result.Err
	}
	if !result.IsHealthy() {
		h.logger.Warn("vault health job: vault is not healthy",
			zap.Bool("sealed", result.Sealed),
			zap.Bool("initialized", result.Initialized),
		)
		h.notify(result)
	}
	return nil
}

// notify calls onSealed if it is set.
func (h *HealthJob) notify(result monitor.HealthResult) {
	if h.onSealed != nil {
		h.onSealed(result)
	}
}
