package monitor

import (
	"fmt"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

type ReplicationMonitor struct {
	checker *vault.ReplicationChecker
	logger  *zap.Logger
}

type ReplicationResult struct {
	DRMode          string
	DRState         string
	PerformanceMode string
	Healthy         bool
	Message         string
}

func NewReplicationMonitor(checker *vault.ReplicationChecker, logger *zap.Logger) (*ReplicationMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &ReplicationMonitor{checker: checker, logger: logger}, nil
}

func (m *ReplicationMonitor) Check() (*ReplicationResult, error) {
	status, err := m.checker.GetReplicationStatus()
	if err != nil {
		m.logger.Error("failed to get replication status", zap.Error(err))
		return nil, fmt.Errorf("replication status check failed: %w", err)
	}

	result := &ReplicationResult{
		DRMode:          status.Data.DR.Mode,
		DRState:         status.Data.DR.State,
		PerformanceMode: status.Data.Performance.Mode,
		Healthy:         true,
		Message:         "replication status OK",
	}

	if status.Data.DR.Mode == "primary" && status.Data.DR.State != "running" {
		result.Healthy = false
		result.Message = fmt.Sprintf("DR replication unhealthy: state=%s", status.Data.DR.State)
		m.logger.Warn("DR replication unhealthy", zap.String("state", status.Data.DR.State))
	}

	if status.Data.Performance.Mode == "primary" && status.Data.Performance.State != "running" {
		result.Healthy = false
		result.Message = fmt.Sprintf("performance replication unhealthy: state=%s", status.Data.Performance.State)
		m.logger.Warn("performance replication unhealthy", zap.String("state", status.Data.Performance.State))
	}

	return result, nil
}
