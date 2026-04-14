package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// CheckResult holds the result of evaluating a single secret's lease.
type CheckResult struct {
	Path     string
	LeaseTTL time.Duration
	Warning  bool
	Expired  bool
	Err      error
}

// SecretChecker checks Vault secrets against configured thresholds.
type SecretChecker struct {
	client          *vault.Client
	warningThreshold time.Duration
	logger          *slog.Logger
}

// NewChecker creates a SecretChecker with the provided client and thresholds.
func NewChecker(client *vault.Client, warningThreshold time.Duration, logger *slog.Logger) *SecretChecker {
	if logger == nil {
		logger = slog.Default()
	}
	return &SecretChecker{
		client:          client,
		warningThreshold: warningThreshold,
		logger:          logger,
	}
}

// CheckSecret fetches the lease for path and evaluates its expiry state.
func (c *SecretChecker) CheckSecret(ctx context.Context, path string) CheckResult {
	lease, err := c.client.GetSecretLease(ctx, path)
	if err != nil {
		return CheckResult{
			Path: path,
			Err:  fmt.Errorf("fetching lease for %s: %w", path, err),
		}
	}

	ttl := time.Duration(lease.LeaseDuration) * time.Second
	result := CheckResult{
		Path:     path,
		LeaseTTL: ttl,
	}

	switch {
	case ttl <= 0:
		result.Expired = true
		c.logger.Warn("secret expired", "path", path)
	case ttl <= c.warningThreshold:
		result.Warning = true
		c.logger.Warn("secret expiring soon", "path", path, "ttl", ttl)
	default:
		c.logger.Info("secret healthy", "path", path, "ttl", ttl)
	}

	return result
}

// CheckAll checks every path and returns a slice of results.
func (c *SecretChecker) CheckAll(ctx context.Context, paths []string) []CheckResult {
	results := make([]CheckResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, c.CheckSecret(ctx, p))
	}
	return results
}

// Summary returns counts of healthy, warning, expired, and errored results
// from the provided slice, useful for logging or metrics reporting.
func Summary(results []CheckResult) (healthy, warning, expired, errored int) {
	for _, r := range results {
		switch {
		case r.Err != nil:
			errored++
		case r.Expired:
			expired++
		case r.Warning:
			warning++
		default:
			healthy++
		}
	}
	return
}
