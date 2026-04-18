package monitor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/user/vaultwatch/internal/vault"
)

type NamespaceMonitor struct {
	checker *vault.NamespaceChecker
	logger  *zap.Logger
}

type NamespaceResult struct {
	Count     int
	Namespaces []vault.NamespaceInfo
	Error     error
}

func NewNamespaceMonitor(checker *vault.NamespaceChecker, logger *zap.Logger) (*NamespaceMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &NamespaceMonitor{checker: checker, logger: logger}, nil
}

func (m *NamespaceMonitor) Check(ctx context.Context) NamespaceResult {
	namespaces, err := m.checker.ListNamespaces(ctx)
	if err != nil {
		m.logger.Warn("failed to list namespaces", zap.Error(err))
		return NamespaceResult{Error: err}
	}

	m.logger.Info("namespace check complete",
		zap.Int("count", len(namespaces)),
	)

	return NamespaceResult{
		Count:      len(namespaces),
		Namespaces: namespaces,
	}
}
