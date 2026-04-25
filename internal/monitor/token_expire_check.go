package monitor

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// TokenExpireMonitor monitors a token's expiry and returns a CheckResult.
type TokenExpireMonitor struct {
	checker       *vault.TokenExpireChecker
	token         string
	warnThreshold time.Duration
	logger        *zap.Logger
}

// NewTokenExpireMonitor creates a new TokenExpireMonitor.
func NewTokenExpireMonitor(checker *vault.TokenExpireChecker, token string, warnThreshold time.Duration, logger *zap.Logger) (*TokenExpireMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	if token == "" {
		return nil, fmt.Errorf("token must not be empty")
	}
	if warnThreshold <= 0 {
		warnThreshold = 24 * time.Hour
	}
	return &TokenExpireMonitor{
		checker:       checker,
		token:         token,
		warnThreshold: warnThreshold,
		logger:        logger,
	}, nil
}

// Check evaluates the token expiry and returns a CheckResult.
func (m *TokenExpireMonitor) Check(ctx context.Context) CheckResult {
	info, err := m.checker.GetTokenExpiry(ctx, m.token)
	if err != nil {
		m.logger.Error("token expire check failed", zap.Error(err))
		return CheckResult{
			Path:    "token:" + m.token[:min(8, len(m.token))],
			Status:  StatusExpired,
			Message: fmt.Sprintf("failed to retrieve token expiry: %v", err),
		}
	}

	ttl := time.Until(info.ExpireTime)
	status := StatusOK
	msg := fmt.Sprintf("token expires in %s", ttl.Round(time.Second))

	switch {
	case ttl <= 0:
		status = StatusExpired
		msg = "token has expired"
	case ttl <= m.warnThreshold:
		status = StatusWarning
		msg = fmt.Sprintf("token expires soon: %s remaining", ttl.Round(time.Second))
	}

	m.logger.Info("token expire check",
		zap.String("token_id", info.TokenID),
		zap.String("status", string(status)),
		zap.Duration("ttl", ttl),
	)

	return CheckResult{
		Path:    "token:" + info.TokenID,
		Status:  status,
		Message: msg,
		TTL:     ttl,
		Renewable: info.Renewable,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
