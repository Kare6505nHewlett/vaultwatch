package monitor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

type EngineMonitor struct {
	checker       *vault.EngineChecker
	expectedTypes []string
	logger        *zap.Logger
}

type EngineResult struct {
	Path    string
	Type    string
	Healthy bool
	Message string
}

func NewEngineMonitor(checker *vault.EngineChecker, expectedTypes []string, logger *zap.Logger) (*EngineMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &EngineMonitor{checker: checker, expectedTypes: expectedTypes, logger: logger}, nil
}

func (m *EngineMonitor) Check(ctx context.Context) ([]EngineResult, error) {
	engines, err := m.checker.ListEngines(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing engines: %w", err)
	}

	present := make(map[string]string, len(engines))
	for _, e := range engines {
		present[e.Path] = e.Type
	}

	var results []EngineResult
	for _, expected := range m.expectedTypes {
		if typ, ok := present[expected]; ok {
			results = append(results, EngineResult{
				Path:    expected,
				Type:    typ,
				Healthy: true,
				Message: "engine mounted",
			})
		} else {
			m.logger.Warn("expected engine not mounted", zap.String("path", expected))
			results = append(results, EngineResult{
				Path:    expected,
				Healthy: false,
				Message: "engine not found",
			})
		}
	}
	return results, nil
}
