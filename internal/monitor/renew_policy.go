package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

// RenewPolicy defines when and how secrets should be auto-renewed.
type RenewPolicy struct {
	// RenewBefore specifies how far before expiry to trigger renewal.
	RenewBefore time.Duration
	// Increment is the requested new lease duration.
	Increment time.Duration
}

// DefaultRenewPolicy returns a sensible default renewal policy.
func DefaultRenewPolicy() RenewPolicy {
	return RenewPolicy{
		RenewBefore: 10 * time.Minute,
		Increment:   1 * time.Hour,
	}
}

// RenewIfNeeded checks a CheckResult and renews the lease if it falls within
// the renewal window defined by the policy. Returns the RenewResult or nil
// if renewal was not attempted.
func RenewIfNeeded(
	ctx context.Context,
	result CheckResult,
	policy RenewPolicy,
	renewer *vault.Renewer,
	logger *zap.Logger,
) (*vault.RenewResult, error) {
	if renewer == nil {
		return nil, fmt.Errorf("renewer must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}

	if result.TTL > policy.RenewBefore {
		logger.Debug("renewal not needed",
			zap.String("path", result.Path),
			zap.Duration("ttl", result.TTL),
			zap.Duration("renew_before", policy.RenewBefore),
		)
		return nil, nil
	}

	logger.Info("secret within renewal window, renewing",
		zap.String("path", result.Path),
		zap.Duration("ttl", result.TTL),
	)

	rrr := renewer.RenewLease(ctx, result.Path, policy.Increment)
	return &rrr, rrr.Error
}
