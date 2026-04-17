package monitor

import (
	"fmt"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

// AuditCheckResult holds the result of an audit device check.
type AuditCheckResult struct {
	DeviceCount int
	Devices     []vault.AuditDevice
	Warning     string
}

// AuditMonitor checks that at least one audit device is enabled.
type AuditMonitor struct {
	checker *vault.AuditChecker
	logger  *zap.Logger
}

// NewAuditMonitor returns a new AuditMonitor or an error if dependencies are nil.
func NewAuditMonitor(checker *vault.AuditChecker, logger *zap.Logger) (*AuditMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("audit checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &AuditMonitor{checker: checker, logger: logger}, nil
}

// Check lists audit devices and warns if none are enabled.
func (a *AuditMonitor) Check() (*AuditCheckResult, error) {
	devices, err := a.checker.ListAuditDevices()
	if err != nil {
		return nil, fmt.Errorf("listing audit devices: %w", err)
	}

	result := &AuditCheckResult{
		DeviceCount: len(devices),
		Devices:     devices,
	}

	if len(devices) == 0 {
		result.Warning = "no audit devices enabled — audit logging is disabled"
		a.logger.Warn("no audit devices enabled")
	} else {
		a.logger.Info("audit devices enabled", zap.Int("count", len(devices)))
	}

	return result, nil
}
