package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// RotatePolicy defines when automatic rotation should be triggered.
type RotatePolicy struct {
	// RotateWithin triggers rotation when TTL falls below this threshold.
	RotateWithin time.Duration
}

// DefaultRotatePolicy returns a sensible default rotation policy.
func DefaultRotatePolicy() RotatePolicy {
	return RotatePolicy{
		RotateWithin: 24 * time.Hour,
	}
}

// RotateIfNeeded checks the CheckResult and triggers rotation via the Rotator
// if the remaining TTL is within the policy threshold.
func RotateIfNeeded(
	ctx context.Context,
	result CheckResult,
	policy RotatePolicy,
	rotator *vault.Rotator,
	logger *slog.Logger,
) error {
	if rotator == nil {
		return fmt.Errorf("rotator must not be nil")
	}
	if logger == nil {
		return fmt.Errorf("logger must not be nil")
	}

	if result.TTL <= 0 || result.TTL > policy.RotateWithin {
		logger.Debug("rotation not needed", "path", result.Path, "ttl", result.TTL)
		return nil
	}

	logger.Info("TTL within rotation threshold, rotating",
		"path", result.Path,
		"ttl", result.TTL,
		"threshold", policy.RotateWithin,
	)

	_, err := rotator.RotateSecret(ctx, result.Path)
	if err != nil {
		return fmt.Errorf("rotate %s: %w", result.Path, err)
	}
	return nil
}
