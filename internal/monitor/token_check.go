package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// TokenChecker checks the expiry status of a Vault token.
type TokenChecker struct {
	client *vault.Client
	logger *slog.Logger
	warnThreshold time.Duration
}

// NewTokenChecker creates a new TokenChecker.
func NewTokenChecker(client *vault.Client, logger *slog.Logger, warnThreshold time.Duration) (*TokenChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	if warnThreshold <= 0 {
		warnThreshold = 24 * time.Hour
	}
	return &TokenChecker{
		client:        client,
		logger:        logger,
		warnThreshold: warnThreshold,
	}, nil
}

// CheckToken looks up the current Vault token and returns a CheckResult.
func (tc *TokenChecker) CheckToken(ctx context.Context) (CheckResult, error) {
	info, err := tc.client.GetTokenInfo(ctx)
	if err != nil {
		return CheckResult{}, fmt.Errorf("failed to get token info: %w", err)
	}

	now := time.Now()
	result := CheckResult{
		Path:      "auth/token/lookup-self",
		LeaseTTL:  info.TTL,
		ExpiresAt: now.Add(info.TTL),
	}

	switch {
	case info.TTL <= 0:
		result.Status = StatusExpired
		tc.logger.Warn("vault token has expired")
	case info.TTL <= tc.warnThreshold:
		result.Status = StatusWarning
		tc.logger.Warn("vault token expiring soon", "ttl", info.TTL)
	default:
		result.Status = StatusOK
		tc.logger.Info("vault token is healthy", "ttl", info.TTL)
	}

	return result, nil
}

// IsExpired returns true if the token is expired or will expire within the given duration.
func (tc *TokenChecker) IsExpired(ctx context.Context, within time.Duration) (bool, error) {
	result, err := tc.CheckToken(ctx)
	if err != nil {
		return false, err
	}
	return result.Status == StatusExpired || result.LeaseTTL <= within, nil
}
