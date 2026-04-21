package monitor

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// TokenRenewer is the interface for renewing Vault tokens.
type TokenRenewer interface {
	RenewSelf(ctx context.Context, increment int) (interface{ GetLeaseDuration() int }, error)
}

// TokenRenewPolicy defines when a token should be renewed.
type TokenRenewPolicy struct {
	RenewBeforeExpiry time.Duration
	Increment         int
}

// DefaultTokenRenewPolicy returns a policy that renews tokens with 24h remaining.
func DefaultTokenRenewPolicy() TokenRenewPolicy {
	return TokenRenewPolicy{
		RenewBeforeExpiry: 24 * time.Hour,
		Increment:         0,
	}
}

// RenewTokenIfNeeded renews the token if its TTL is within the policy threshold.
// ttl is the current remaining TTL of the token.
func RenewTokenIfNeeded(ctx context.Context, renewer interface {
	RenewSelf(ctx context.Context, increment int) error
}, ttl time.Duration, policy TokenRenewPolicy, logger *zap.Logger) error {
	if ttl <= 0 {
		return fmt.Errorf("token has already expired (ttl=%s)", ttl)
	}
	if ttl > policy.RenewBeforeExpiry {
		logger.Debug("token renewal not needed",
			zap.Duration("ttl", ttl),
			zap.Duration("threshold", policy.RenewBeforeExpiry),
		)
		return nil
	}

	logger.Info("renewing token",
		zap.Duration("ttl", ttl),
		zap.Duration("threshold", policy.RenewBeforeExpiry),
	)

	if err := renewer.RenewSelf(ctx, policy.Increment); err != nil {
		return fmt.Errorf("token renewal failed: %w", err)
	}

	logger.Info("token renewed successfully")
	return nil
}
