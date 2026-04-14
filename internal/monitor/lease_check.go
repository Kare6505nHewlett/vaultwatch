package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

// LeaseCheckResult holds the outcome of a single lease check.
type LeaseCheckResult struct {
	LeaseID   string
	TTL       time.Duration
	ExpiresAt time.Time
	Renewable bool
	Status    string // "ok", "warning", "expired"
	Error     error
}

// LeaseChecker checks lease expiry using a LeaseManager.
type LeaseChecker struct {
	manager       *vault.LeaseManager
	warningWindow time.Duration
	logger        *zap.Logger
}

// NewLeaseChecker creates a LeaseChecker with the given warning window.
func NewLeaseChecker(manager *vault.LeaseManager, warningWindow time.Duration, logger *zap.Logger) (*LeaseChecker, error) {
	if manager == nil {
		return nil, fmt.Errorf("lease manager must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	if warningWindow <= 0 {
		warningWindow = 24 * time.Hour
	}
	return &LeaseChecker{manager: manager, warningWindow: warningWindow, logger: logger}, nil
}

// Check evaluates the expiry status of the given lease ID.
func (lc *LeaseChecker) Check(ctx context.Context, leaseID string) LeaseCheckResult {
	result := LeaseCheckResult{LeaseID: leaseID}

	info, err := lc.manager.GetLeaseInfo(ctx, leaseID)
	if err != nil {
		result.Error = err
		result.Status = "expired"
		lc.logger.Error("failed to get lease info", zap.String("lease_id", leaseID), zap.Error(err))
		return result
	}

	result.TTL = info.TTL
	result.ExpiresAt = info.ExpiresAt
	result.Renewable = info.Renewable

	switch {
	case info.TTL <= 0:
		result.Status = "expired"
	case info.TTL <= lc.warningWindow:
		result.Status = "warning"
	default:
		result.Status = "ok"
	}

	lc.logger.Info("lease check complete",
		zap.String("lease_id", leaseID),
		zap.String("status", result.Status),
		zap.Duration("ttl", info.TTL),
	)
	return result
}
