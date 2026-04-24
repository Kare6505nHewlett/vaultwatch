package monitor

import (
	"fmt"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

// TokenAccessorMonitor checks token health via accessor lookup.
type TokenAccessorMonitor struct {
	checker   *vault.TokenAccessorChecker
	accessors []string
	warnTTL   time.Duration
	logger    *zap.Logger
}

// NewTokenAccessorMonitor creates a TokenAccessorMonitor.
func NewTokenAccessorMonitor(checker *vault.TokenAccessorChecker, accessors []string, warnTTL time.Duration, logger *zap.Logger) (*TokenAccessorMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("token accessor checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	if warnTTL <= 0 {
		warnTTL = 24 * time.Hour
	}
	return &TokenAccessorMonitor{
		checker:   checker,
		accessors: accessors,
		warnTTL:   warnTTL,
		logger:    logger,
	}, nil
}

// CheckResult holds the result for a single accessor check.
type AccessorCheckResult struct {
	Accessor    string
	DisplayName string
	TTL         time.Duration
	Status      string
	Message     string
}

// Check evaluates all configured accessors and returns results.
func (m *TokenAccessorMonitor) Check() []AccessorCheckResult {
	results := make([]AccessorCheckResult, 0, len(m.accessors))
	for _, acc := range m.accessors {
		info, err := m.checker.LookupByAccessor(acc)
		if err != nil {
			m.logger.Warn("failed to lookup accessor", zap.String("accessor", acc), zap.Error(err))
			results = append(results, AccessorCheckResult{
				Accessor: acc,
				Status:  "error",
				Message: err.Error(),
			})
			continue
		}
		ttl := time.Duration(info.TTL) * time.Second
		status := "ok"
		msg := fmt.Sprintf("TTL: %s", ttl.Round(time.Second))
		if ttl <= 0 {
			status = "expired"
			msg = "token has expired"
		} else if ttl <= m.warnTTL {
			status = "warning"
			msg = fmt.Sprintf("token expires soon: %s remaining", ttl.Round(time.Second))
		}
		m.logger.Info("accessor check", zap.String("accessor", acc), zap.String("status", status))
		results = append(results, AccessorCheckResult{
			Accessor:    acc,
			DisplayName: info.DisplayName,
			TTL:         ttl,
			Status:      status,
			Message:     msg,
		})
	}
	return results
}
