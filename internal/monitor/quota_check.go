package monitor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

type QuotaMonitor struct {
	checker *vault.QuotaChecker
	logger  *zap.Logger
}

type QuotaResult struct {
	TotalQuotas int
	Names       []string
	Healthy     bool
	Message     string
}

func NewQuotaMonitor(checker *vault.QuotaChecker, logger *zap.Logger) (*QuotaMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &QuotaMonitor{checker: checker, logger: logger}, nil
}

func (m *QuotaMonitor) Check(ctx context.Context) (*QuotaResult, error) {
	quotas, err := m.checker.ListQuotas(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing quotas: %w", err)
	}

	names := make([]string, 0, len(quotas))
	for _, q := range quotas {
		names = append(names, q.Name)
	}

	result := &QuotaResult{
		TotalQuotas: len(quotas),
		Names:       names,
		Healthy:     true,
		Message:     fmt.Sprintf("%d rate-limit quota(s) configured", len(quotas)),
	}

	m.logger.Info("quota check complete", zap.Int("total", result.TotalQuotas))
	return result, nil
}
