package vault

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// RenewResult holds the outcome of a lease or token renewal attempt.
type RenewResult struct {
	Path      string
	Renewed   bool
	NewExpiry time.Time
	Error     error
}

// Renewer wraps a Client and provides lease/token renewal capabilities.
type Renewer struct {
	client *Client
	logger *zap.Logger
}

// NewRenewer creates a new Renewer using the provided Client and logger.
func NewRenewer(client *Client, logger *zap.Logger) (*Renewer, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &Renewer{client: client, logger: logger}, nil
}

// RenewLease attempts to renew the lease at the given path and returns a RenewResult.
func (r *Renewer) RenewLease(ctx context.Context, path string, increment time.Duration) RenewResult {
	r.logger.Info("attempting lease renewal", zap.String("path", path))

	data := map[string]interface{}{
		"increment": int(increment.Seconds()),
	}

	secret, err := r.client.logical.WriteWithContext(ctx, "sys/leases/renew", data)
	if err != nil {
		r.logger.Warn("lease renewal failed", zap.String("path", path), zap.Error(err))
		return RenewResult{Path: path, Renewed: false, Error: err}
	}

	if secret == nil {
		err = fmt.Errorf("empty response from Vault for path %s", path)
		return RenewResult{Path: path, Renewed: false, Error: err}
	}

	newExpiry := time.Now().Add(time.Duration(secret.LeaseDuration) * time.Second)
	r.logger.Info("lease renewed successfully",
		zap.String("path", path),
		zap.Time("new_expiry", newExpiry),
	)

	return RenewResult{
		Path:      path,
		Renewed:   true,
		NewExpiry: newExpiry,
	}
}
