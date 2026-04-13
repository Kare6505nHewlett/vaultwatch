package monitor

import (
	"fmt"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// ExpiryStatus represents the expiry state of a secret lease.
type ExpiryStatus struct {
	Path        string
	LeaseDuration int
	Renewable   bool
	ExpiresAt   time.Time
	Warning     bool
	Expired     bool
	Message     string
}

// Checker holds configuration for expiry checking.
type Checker struct {
	client        *vault.Client
	warnThreshold time.Duration
}

// NewChecker creates a new Checker with the given Vault client and warning threshold.
func NewChecker(client *vault.Client, warnThreshold time.Duration) *Checker {
	return &Checker{
		client:        client,
		warnThreshold: warnThreshold,
	}
}

// CheckSecret retrieves the lease info for a secret path and evaluates its expiry status.
func (c *Checker) CheckSecret(path string) (*ExpiryStatus, error) {
	lease, err := c.client.GetSecretLease(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get lease for %q: %w", path, err)
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(lease.LeaseDuration) * time.Second)

	status := &ExpiryStatus{
		Path:          path,
		LeaseDuration: lease.LeaseDuration,
		Renewable:     lease.Renewable,
		ExpiresAt:     expiresAt,
	}

	switch {
	case lease.LeaseDuration <= 0:
		status.Expired = true
		status.Message = fmt.Sprintf("secret %q has expired or has no lease", path)
	case expiresAt.Before(now.Add(c.warnThreshold)):
		status.Warning = true
		status.Message = fmt.Sprintf("secret %q expires in %s", path, time.Until(expiresAt).Round(time.Second))
	default:
		status.Message = fmt.Sprintf("secret %q is valid, expires at %s", path, expiresAt.Format(time.RFC3339))
	}

	return status, nil
}

// CheckSecrets checks multiple secret paths and returns all statuses.
func (c *Checker) CheckSecrets(paths []string) ([]*ExpiryStatus, error) {
	var statuses []*ExpiryStatus
	for _, path := range paths {
		status, err := c.CheckSecret(path)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}
