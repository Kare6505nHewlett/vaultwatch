package monitor

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// TokenMetadataMonitor checks that a token accessor resolves and has expected policies.
type TokenMetadataMonitor struct {
	checker          *vault.TokenMetadataChecker
	accessor         string
	requiredPolicies []string
	logger           *zap.Logger
}

// TokenMetadataResult holds the result of a token metadata check.
type TokenMetadataResult struct {
	Accessor        string
	DisplayName     string
	Policies        []string
	MissingPolicies []string
	Healthy         bool
	Message         string
}

// NewTokenMetadataMonitor creates a new TokenMetadataMonitor.
func NewTokenMetadataMonitor(checker *vault.TokenMetadataChecker, accessor string, required []string, logger *zap.Logger) (*TokenMetadataMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	if accessor == "" {
		return nil, fmt.Errorf("accessor must not be empty")
	}
	return &TokenMetadataMonitor{
		checker:          checker,
		accessor:         accessor,
		requiredPolicies: required,
		logger:           logger,
	}, nil
}

// Check retrieves token metadata and validates required policies.
func (m *TokenMetadataMonitor) Check() (*TokenMetadataResult, error) {
	meta, err := m.checker.GetTokenMetadata(m.accessor)
	if err != nil {
		return nil, fmt.Errorf("fetching token metadata: %w", err)
	}

	policySet := make(map[string]struct{}, len(meta.Policies))
	for _, p := range meta.Policies {
		policySet[p] = struct{}{}
	}

	var missing []string
	for _, req := range m.requiredPolicies {
		if _, ok := policySet[req]; !ok {
			missing = append(missing, req)
		}
	}

	healthy := len(missing) == 0
	msg := "token metadata OK"
	if !healthy {
		msg = fmt.Sprintf("missing required policies: %v", missing)
	}

	m.logger.Info("token metadata check",
		zap.String("accessor", m.accessor),
		zap.Bool("healthy", healthy),
		zap.Strings("missing_policies", missing),
	)

	return &TokenMetadataResult{
		Accessor:        m.accessor,
		DisplayName:     meta.DisplayName,
		Policies:        meta.Policies,
		MissingPolicies: missing,
		Healthy:         healthy,
		Message:         msg,
	}, nil
}
