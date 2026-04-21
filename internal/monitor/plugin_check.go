package monitor

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// PluginMonitor checks that expected plugins are registered in Vault.
type PluginMonitor struct {
	checker        *vault.PluginChecker
	logger         *zap.Logger
	expectedPlugin []string
	pluginType     string
}

// PluginCheckResult holds the result of a plugin presence check.
type PluginCheckResult struct {
	PluginType string
	Missing    []string
	Present    []string
	Healthy    bool
}

// NewPluginMonitor creates a PluginMonitor that validates expected plugins exist.
func NewPluginMonitor(checker *vault.PluginChecker, logger *zap.Logger, pluginType string, expected []string) (*PluginMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("plugin checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &PluginMonitor{
		checker:        checker,
		logger:         logger,
		expectedPlugin: expected,
		pluginType:     pluginType,
	}, nil
}

// Check fetches registered plugins and reports which expected plugins are missing.
func (m *PluginMonitor) Check(ctx context.Context) (*PluginCheckResult, error) {
	plugins, err := m.checker.ListPlugins(ctx, m.pluginType)
	if err != nil {
		return nil, fmt.Errorf("listing plugins of type %q: %w", m.pluginType, err)
	}

	registered := make(map[string]struct{}, len(plugins))
	for _, p := range plugins {
		registered[p.Name] = struct{}{}
	}

	var missing, present []string
	for _, name := range m.expectedPlugin {
		if _, ok := registered[name]; ok {
			present = append(present, name)
		} else {
			missing = append(missing, name)
			m.logger.Warn("expected plugin not found", zap.String("plugin", name), zap.String("type", m.pluginType))
		}
	}

	result := &PluginCheckResult{
		PluginType: m.pluginType,
		Missing:    missing,
		Present:    present,
		Healthy:    len(missing) == 0,
	}

	m.logger.Info("plugin check complete",
		zap.String("type", m.pluginType),
		zap.Int("present", len(present)),
		zap.Int("missing", len(missing)),
	)
	return result, nil
}
