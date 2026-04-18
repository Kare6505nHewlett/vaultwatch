package vault

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// PolicyMonitor checks that required policies exist in Vault.
type PolicyMonitor struct {
	checker  *PolicyChecker
	logger   *zap.Logger
	policies []string
}

// PolicyResult holds the result of a policy existence check.
type PolicyResult struct {
	Policy  string
	Exists  bool
	Error   error
}

// NewPolicyMonitor creates a PolicyMonitor for the given policy names.
func NewPolicyMonitor(checker *PolicyChecker, logger *zap.Logger, policies []string) (*PolicyMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("policy checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	if len(policies) == 0 {
		return nil, fmt.Errorf("at least one policy name is required")
	}
	return &PolicyMonitor{
		checker:  checker,
		logger:   logger,
		policies: policies,
	}, nil
}

// Check verifies each required policy exists and returns results.
func (m *PolicyMonitor) Check(ctx context.Context) []PolicyResult {
	results := make([]PolicyResult, 0, len(m.policies))
	for _, name := range m.policies {
		_, err := m.checker.GetPolicy(ctx, name)
		result := PolicyResult{Policy: name}
		if err != nil {
			m.logger.Warn("policy check failed", zap.String("policy", name), zap.Error(err))
			result.Exists = false
			result.Error = err
		} else {
			m.logger.Info("policy exists", zap.String("policy", name))
			result.Exists = true
		}
		results = append(results, result)
	}
	return results
}
