package monitor

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// TokenCountMonitor monitors the total number of active tokens against a threshold.
type TokenCountMonitor struct {
	checker   *vault.TokenCountChecker
	logger    *zap.Logger
	threshold int
}

// TokenCountResult holds the outcome of a token count check.
type TokenCountResult struct {
	Total     int
	Threshold int
	Exceeded  bool
	Message   string
}

// NewTokenCountMonitor creates a new TokenCountMonitor.
func NewTokenCountMonitor(checker *vault.TokenCountChecker, threshold int, logger *zap.Logger) (*TokenCountMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("token count checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	if threshold <= 0 {
		threshold = 1000
	}
	return &TokenCountMonitor{checker: checker, logger: logger, threshold: threshold}, nil
}

// Check queries Vault for the current token count and evaluates it against the threshold.
func (m *TokenCountMonitor) Check() (*TokenCountResult, error) {
	result, err := m.checker.GetTokenCount()
	if err != nil {
		return nil, fmt.Errorf("getting token count: %w", err)
	}

	exceeded := result.TotalCount >= m.threshold
	msg := fmt.Sprintf("token count %d is within threshold %d", result.TotalCount, m.threshold)
	if exceeded {
		msg = fmt.Sprintf("token count %d meets or exceeds threshold %d", result.TotalCount, m.threshold)
		m.logger.Warn("token count threshold exceeded",
			zap.Int("total", result.TotalCount),
			zap.Int("threshold", m.threshold),
		)
	} else {
		m.logger.Info("token count OK",
			zap.Int("total", result.TotalCount),
			zap.Int("threshold", m.threshold),
		)
	}

	return &TokenCountResult{
		Total:     result.TotalCount,
		Threshold: m.threshold,
		Exceeded:  exceeded,
		Message:   msg,
	}, nil
}
