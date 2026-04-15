package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

// HealthResult holds the result of a Vault health check.
type HealthResult struct {
	Timestamp   time.Time
	Initialized bool
	Sealed      bool
	Standby     bool
	Version     string
	ClusterName string
	Err         error
}

// HealthMonitor periodically checks Vault server health.
type HealthMonitor struct {
	checker *vault.HealthChecker
	logger  *zap.Logger
}

// NewHealthMonitor creates a new HealthMonitor.
func NewHealthMonitor(checker *vault.HealthChecker, logger *zap.Logger) (*HealthMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("health checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &HealthMonitor{checker: checker, logger: logger}, nil
}

// Run performs a single health check and returns the result.
func (m *HealthMonitor) Run(ctx context.Context) HealthResult {
	result := HealthResult{Timestamp: time.Now()}

	status, err := m.checker.Check(ctx)
	if err != nil {
		result.Err = err
		m.logger.Error("vault health check failed", zap.Error(err))
		return result
	}

	result.Initialized = status.Initialized
	result.Sealed = status.Sealed
	result.Standby = status.Standby
	result.Version = status.Version
	result.ClusterName = status.ClusterName

	if status.Sealed {
		m.logger.Warn("vault is sealed", zap.String("cluster", status.ClusterName))
	} else if !status.Initialized {
		m.logger.Warn("vault is not initialized")
	} else {
		m.logger.Info("vault is healthy",
			zap.String("version", status.Version),
			zap.String("cluster", status.ClusterName),
		)
	}
	return result
}

// IsHealthy returns true if the result indicates a healthy, unsealed, initialized Vault.
func (r HealthResult) IsHealthy() bool {
	return r.Err == nil && r.Initialized && !r.Sealed
}
