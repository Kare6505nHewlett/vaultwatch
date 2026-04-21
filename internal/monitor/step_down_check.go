package monitor

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// StepDownMonitor monitors whether the Vault step-down endpoint is accessible.
type StepDownMonitor struct {
	checker *vault.StepDownChecker
	logger  *zap.Logger
}

// StepDownStatus represents the result of a step-down endpoint check.
type StepDownStatus struct {
	Healthy bool
	Message string
}

// NewStepDownMonitor creates a new StepDownMonitor.
func NewStepDownMonitor(checker *vault.StepDownChecker, logger *zap.Logger) (*StepDownMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("step-down checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &StepDownMonitor{checker: checker, logger: logger}, nil
}

// Check probes the step-down endpoint and returns a StepDownStatus.
func (m *StepDownMonitor) Check(ctx context.Context) (*StepDownStatus, error) {
	result, err := m.checker.CheckStepDownEndpoint(ctx)
	if err != nil {
		return nil, fmt.Errorf("step-down endpoint check failed: %w", err)
	}

	if !result.Reachable {
		m.logger.Warn("step-down endpoint not reachable")
		return &StepDownStatus{Healthy: false, Message: "endpoint unreachable"}, nil
	}

	// 204 No Content is the expected response when the endpoint is available.
	// 403 means the token lacks permission but the endpoint exists.
	healthy := result.StatusCode == http.StatusNoContent ||
		result.StatusCode == http.StatusForbidden

	msg := result.Message
	if !healthy {
		m.logger.Warn("unexpected step-down response", zap.Int("status", result.StatusCode))
		msg = fmt.Sprintf("unexpected status: %d", result.StatusCode)
	}

	return &StepDownStatus{Healthy: healthy, Message: msg}, nil
}
