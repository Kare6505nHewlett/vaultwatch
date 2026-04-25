package monitor

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// TokenRevoker is the interface for revoking tokens.
type TokenRevoker interface {
	RevokeSelf(ctx context.Context) error
	RevokeAccessor(ctx context.Context, accessor string) error
}

// TokenRevokeMonitor monitors and revokes tokens on demand.
type TokenRevokeMonitor struct {
	revoker   TokenRevoker
	accessors []string
	logger    *zap.Logger
}

// NewTokenRevokeMonitor creates a new TokenRevokeMonitor.
func NewTokenRevokeMonitor(revoker TokenRevoker, accessors []string, logger *zap.Logger) (*TokenRevokeMonitor, error) {
	if revoker == nil {
		return nil, fmt.Errorf("revoker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenRevokeMonitor{
		revoker:   revoker,
		accessors: accessors,
		logger:    logger,
	}, nil
}

// RevokeAll revokes all configured accessor tokens and, optionally, self.
func (m *TokenRevokeMonitor) RevokeAll(ctx context.Context, revokeSelf bool) error {
	var errs []error

	for _, acc := range m.accessors {
		if err := m.revoker.RevokeAccessor(ctx, acc); err != nil {
			m.logger.Error("failed to revoke accessor",
				zap.String("accessor", acc),
				zap.Error(err))
			errs = append(errs, err)
			continue
		}
		m.logger.Info("revoked accessor token", zap.String("accessor", acc))
	}

	if revokeSelf {
		if err := m.revoker.RevokeSelf(ctx); err != nil {
			m.logger.Error("failed to revoke self token", zap.Error(err))
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d revocation(s) failed", len(errs))
	}
	return nil
}
