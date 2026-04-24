package scheduler

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// TransitJob is a scheduler job that checks transit key health.
type TransitJob struct {
	monitor *monitor.TransitMonitor
	logger  *log.Logger
}

// NewTransitJob creates a new TransitJob.
func NewTransitJob(m *monitor.TransitMonitor, logger *log.Logger) (*TransitJob, error) {
	if m == nil {
		return nil, fmt.Errorf("transit job: monitor must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("transit job: logger must not be nil")
	}
	return &TransitJob{monitor: m, logger: logger}, nil
}

// Name returns the job identifier.
func (j *TransitJob) Name() string {
	return "transit-key-check"
}

// Run executes the transit key check and logs results.
func (j *TransitJob) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	results := j.monitor.Check()

	unhealthy := 0
	for _, r := range results {
		if r.Healthy {
			j.logger.Printf("[transit] key=%s type=%s version=%d status=ok",
				r.KeyName, r.Type, r.LatestVersion)
		} else {
			j.logger.Printf("[transit] key=%s status=unhealthy message=%s",
				r.KeyName, r.Message)
			unhealthy++
		}
	}

	if unhealthy > 0 {
		return fmt.Errorf("transit job: %d key(s) unhealthy", unhealthy)
	}
	return nil
}
