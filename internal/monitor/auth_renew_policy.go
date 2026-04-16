package monitor

import (
	"context"
	"log/slog"
	"time"
)

// AuthRenewer is the interface for renewing the Vault auth token.
type AuthRenewer interface {
	RenewSelf(ctx context.Context, increment int) (int, error)
}

// DefaultAuthRenewPolicy defines when a token should be renewed.
type DefaultAuthRenewPolicy struct {
	WarnThreshold time.Duration
	Increment     int
}

// RenewAuthIfNeeded renews the token when remaining TTL is below the warn threshold.
func RenewAuthIfNeeded(ctx context.Context, renewer AuthRenewer, ttl time.Duration, policy DefaultAuthRenewPolicy, logger *slog.Logger) (bool, error) {
	if ttl > policy.WarnThreshold {
		logger.DebugContext(ctx, "token renewal not needed", "ttl", ttl)
		return false, nil
	}

	increment := policy.Increment
	if increment == 0 {
		increment = int(policy.WarnThreshold.Seconds()) * 2
	}

	newTTL, err := renewer.RenewSelf(ctx, increment)
	if err != nil {
		return false, err
	}

	logger.InfoContext(ctx, "auth token renewed",
		"previous_ttl", ttl,
		"new_ttl_seconds", newTTL,
	)
	return true, nil
}
