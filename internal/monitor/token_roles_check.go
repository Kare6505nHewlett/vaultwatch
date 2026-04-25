package monitor

import (
	"fmt"
	"log"

	"github.com/your-org/vaultwatch/internal/vault"
)

// TokenRolesResult holds the outcome of checking one or more token roles.
type TokenRolesResult struct {
	RoleName       string
	Orphan         bool
	Renewable      bool
	MaxTTL         int
	ExplicitMaxTTL int
	Warning        string
	Healthy        bool
}

// TokenRolesMonitor checks that required token roles exist and meet policy.
type TokenRolesMonitor struct {
	checker       *vault.TokenRolesChecker
	logger        *log.Logger
	requiredRoles []string
	warnMaxTTL    int // warn if token_max_ttl is 0 (unlimited)
}

// NewTokenRolesMonitor creates a new TokenRolesMonitor.
func NewTokenRolesMonitor(checker *vault.TokenRolesChecker, roles []string, logger *log.Logger) (*TokenRolesMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("token roles checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenRolesMonitor{
		checker:       checker,
		logger:        logger,
		requiredRoles: roles,
	}, nil
}

// Check verifies each required token role exists and returns results.
func (m *TokenRolesMonitor) Check() []TokenRolesResult {
	results := make([]TokenRolesResult, 0, len(m.requiredRoles))
	for _, name := range m.requiredRoles {
		role, err := m.checker.GetTokenRole(name)
		if err != nil {
			m.logger.Printf("[token_roles] error fetching role %q: %v", name, err)
			results = append(results, TokenRolesResult{
				RoleName: name,
				Warning:  fmt.Sprintf("role not found or unreachable: %v", err),
				Healthy:  false,
			})
			continue
		}

		res := TokenRolesResult{
			RoleName:       role.Name,
			Orphan:         role.Orphan,
			Renewable:      role.Renewable,
			MaxTTL:         role.MaxTTL,
			ExplicitMaxTTL: role.ExplicitMaxTTL,
			Healthy:        true,
		}
		if role.MaxTTL == 0 {
			res.Warning = fmt.Sprintf("role %q has no max TTL set (unlimited)", name)
			m.logger.Printf("[token_roles] warning: %s", res.Warning)
		}
		results = append(results, res)
	}
	return results
}
