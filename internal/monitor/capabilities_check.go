package monitor

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

// CapabilityRequirement defines a path and the capabilities expected on it.
type CapabilityRequirement struct {
	Path                 string
	RequiredCapabilities []string
}

// CapabilitiesMonitor checks that the current token has required capabilities on configured paths.
type CapabilitiesMonitor struct {
	checker      *vault.CapabilitiesChecker
	requirements []CapabilityRequirement
	logger       *zap.Logger
}

// NewCapabilitiesMonitor creates a new CapabilitiesMonitor.
func NewCapabilitiesMonitor(checker *vault.CapabilitiesChecker, requirements logger *zap.Logger) (*CapabilitiesMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("capabilities checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &CapabilitiesMonitor{
		checker:      checker,
		requirements: requirements,
		logger:       logger,
	}, nil
}

// CapabilityCheckResult holds the result of checking a single path.
type CapabilityCheckResult struct {
	Path    string
	Missing []string
	Healthy bool
}

// Check verifies that the token has all required capabilities for each configured path.
func (m *CapabilitiesMonitor) Check(ctx context.Context) ([]CapabilityCheckResult, error) {
	var results []CapabilityCheckResult

	for _, req := range m.requirements {
		res, err := m.checker.CheckCapabilities(ctx, req.Path)
		if err != nil {
			m.logger.Error("failed to check capabilities", zap.String("path", req.Path), zap.Error(err))
			return nil, fmt.Errorf("capabilities check error for path %q: %w", req.Path, err)
		}

		capSet := make(map[string]struct{}, len(res.Capabilities))
		for _, c := range res.Capabilities {
			capSet[c] = struct{}{}
		}

		var missing []string
		for _, required := range req.RequiredCapabilities {
			if _, ok := capSet[required]; !ok {
				missing = append(missing, required)
			}
		}

		healthy := len(missing) == 0
		if !healthy {
			m.logger.Warn("token missing capabilities", zap.String("path", req.Path), zap.Strings("missing", missing))
		} else {
			m.logger.Debug("capabilities satisfied", zap.String("path", req.Path))
		}

		results = append(results, CapabilityCheckResult{
			Path:    req.Path,
			Missing: missing,
			Healthy: healthy,
		})
	}

	return results, nil
}
