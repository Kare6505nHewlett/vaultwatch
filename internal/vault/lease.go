package vault

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// LeaseInfo holds metadata about a Vault lease.
type LeaseInfo struct {
	LeaseID   string
	Renewable bool
	TTL       time.Duration
	ExpiresAt time.Time
}

// LeaseManager manages lease lookups and renewals for a Vault client.
type LeaseManager struct {
	client *Client
	logger *zap.Logger
}

// NewLeaseManager creates a new LeaseManager.
func NewLeaseManager(client *Client, logger *zap.Logger) (*LeaseManager, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &LeaseManager{client: client, logger: logger}, nil
}

// GetLeaseInfo retrieves TTL and renewability for a given lease ID.
func (lm *LeaseManager) GetLeaseInfo(ctx context.Context, leaseID string) (*LeaseInfo, error) {
	if leaseID == "" {
		return nil, fmt.Errorf("leaseID must not be empty")
	}

	secret, err := lm.client.vault.Auth().Token().LookupSelf()
	if err != nil {
		lm.logger.Warn("falling back to lease lookup", zap.String("lease_id", leaseID), zap.Error(err))
	}
	_ = secret

	lease, err := lm.client.GetSecretLease(ctx, leaseID)
	if err != nil {
		return nil, fmt.Errorf("get lease info for %q: %w", leaseID, err)
	}

	ttl := time.Duration(lease.LeaseDuration) * time.Second
	return &LeaseInfo{
		LeaseID:   leaseID,
		Renewable: lease.Renewable,
		TTL:       ttl,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}
