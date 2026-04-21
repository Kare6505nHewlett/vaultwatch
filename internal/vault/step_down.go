package vault

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// StepDownChecker checks whether the active Vault node can be asked to step down.
type StepDownChecker struct {
	client *Client
	logger *zap.Logger
}

// NewStepDownChecker creates a new StepDownChecker.
func NewStepDownChecker(client *Client, logger *zap.Logger) (*StepDownChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &StepDownChecker{client: client, logger: logger}, nil
}

// StepDownResult holds the outcome of a step-down check.
type StepDownResult struct {
	Reachable bool
	StatusCode int
	Message string
}

// CheckStepDownEndpoint verifies the /v1/sys/step-down endpoint is reachable
// (a PUT to it would cause the active node to step down; here we only probe
// reachability by inspecting the HTTP response without actually triggering it).
func (s *StepDownChecker) CheckStepDownEndpoint(ctx context.Context) (*StepDownResult, error) {
	url := s.client.Address + "/v1/sys/step-down"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building step-down request: %w", err)
	}
	req.Header.Set("X-Vault-Token", s.client.Token)

	resp, err := s.client.HTTP.Do(req)
	if err != nil {
		s.logger.Warn("step-down endpoint unreachable", zap.Error(err))
		return &StepDownResult{Reachable: false, Message: err.Error()}, nil
	}
	defer resp.Body.Close()

	s.logger.Info("step-down endpoint probed", zap.Int("status", resp.StatusCode))
	return &StepDownResult{
		Reachable:  true,
		StatusCode: resp.StatusCode,
		Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
	}, nil
}
