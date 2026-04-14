package monitor

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// KVChecker checks KV v2 secret metadata for deletion or destruction.
type KVChecker struct {
	client  *vault.Client
	logger  *zap.Logger
	warning time.Duration
}

// NewKVChecker creates a KVChecker with the given warning threshold.
func NewKVChecker(client *vault.Client, warning time.Duration, logger *zap.Logger) (*KVChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &KVChecker{
		client:  client,
		logger:  logger,
		warning: warning,
	}, nil
}

// CheckKVSecret looks up KV v2 metadata and returns a CheckResult.
// It reports StatusExpired if the secret version is destroyed or already deleted,
// StatusWarning if deletion is scheduled within the warning threshold, and
// StatusOK otherwise.
func (k *KVChecker) CheckKVSecret(ctx context.Context, mount, path string) CheckResult {
	fullPath := fmt.Sprintf("%s/%s", mount, path)

	meta, err := k.client.LookupSecret(ctx, mount, path)
	if err != nil {
		k.logger.Warn("failed to look up KV secret", zap.String("path", fullPath), zap.Error(err))
		return CheckResult{
			Path:    fullPath,
			Status:  StatusExpired,
			Message: fmt.Sprintf("lookup error: %v", err),
		}
	}

	if meta.Destroyed {
		return CheckResult{
			Path:    fullPath,
			Status:  StatusExpired,
			Message: fmt.Sprintf("version %d is destroyed", meta.Version),
		}
	}

	if meta.DeletedTime != nil {
		until := time.Until(*meta.DeletedTime)
		if until <= 0 {
			return CheckResult{
				Path:    fullPath,
				Status:  StatusExpired,
				TTL:     0,
				Message: fmt.Sprintf("version %d has been deleted", meta.Version),
			}
		}
		status := StatusOK
		if until <= k.warning {
			status = StatusWarning
		}
		return CheckResult{
			Path:    fullPath,
			Status:  status,
			TTL:     until,
			Message: fmt.Sprintf("version %d scheduled for deletion", meta.Version),
		}
	}

	k.logger.Debug("KV secret OK", zap.String("path", fullPath), zap.Int("version", meta.Version))
	return CheckResult{
		Path:    fullPath,
		Status:  StatusOK,
		Message: fmt.Sprintf("version %d is current", meta.Version),
	}
}

// CheckKVSecrets checks multiple KV secrets and returns all results.
// Checks continue even if individual secrets fail.
func (k *KVChecker) CheckKVSecrets(ctx context.Context, mount string, paths []string) []CheckResult {
	results := make([]CheckResult, 0, len(paths))
	for _, path := range paths {
		results = append(results, k.CheckKVSecret(ctx, mount, path))
	}
	return results
}
