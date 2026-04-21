package scheduler

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// StepDownJob is a scheduler job that periodically checks the Vault step-down endpoint.
type StepDownJob struct {
	mon    *monitor.StepDownMonitor
	logger *zap.Logger
}

// NewStepDownJob creates a new StepDownJob.
func NewStepDownJob(mon *monitor.StepDownMonitor, logger *zap.Logger) (*StepDownJob, error) {
	if mon == nil {
		return nil, fmt.Errorf("step-down monitor must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &StepDownJob{mon: mon, logger: logger}, nil
}

// Name returns the job identifier.
func (j *StepDownJob) Name() string {
	return "step-down-check"
}

// Run executes the step-down endpoint check and logs the result.
func (j *StepDownJob) Run(ctx context.Context) error {
	status, err := j.mon.Check(ctx)
	if err != nil {
		j.logger.Error("step-down check error", zap.Error(err))
		return err
	}

	if status.Healthy {
		j.logger.Info("step-down endpoint healthy", zap.String("message", status.Message))
	} else {
		j.logger.Warn("step-down endpoint unhealthy", zap.String("message", status.Message))
	}
	return nil
}
