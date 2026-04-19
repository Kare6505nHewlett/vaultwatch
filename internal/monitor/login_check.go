package monitor

import (
	"fmt"

	"github.com/yourusername/vaultwatch/internal/vault"
	"go.uber.org/zap"
)

type AppRoleLoginMonitor struct {
	checker  *vault.LoginChecker
	roleID   string
	secretID string
	logger   *zap.Logger
}

type LoginCheckResult struct {
	Success     bool
	ClientToken string
	TTL         int
	Renewable   bool
	Message     string
}

func NewAppRoleLoginMonitor(checker *vault.LoginChecker, roleID, secretID string, logger *zap.Logger) (*AppRoleLoginMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	if roleID == "" || secretID == "" {
		return nil, fmt.Errorf("roleID and secretID must not be empty")
	}
	return &AppRoleLoginMonitor{
		checker:  checker,
		roleID:   roleID,
		secretID: secretID,
		logger:   logger,
	}, nil
}

func (m *AppRoleLoginMonitor) Check() LoginCheckResult {
	result, err := m.checker.LoginWithAppRole(m.roleID, m.secretID)
	if err != nil {
		m.logger.Warn("AppRole login check failed", zap.Error(err))
		return LoginCheckResult{
			Success: false,
			Message: fmt.Sprintf("login failed: %v", err),
		}
	}
	m.logger.Info("AppRole login check succeeded",
		zap.String("accessor", result.Accessor),
		zap.Int("ttl", result.TTL),
	)
	return LoginCheckResult{
		Success:     true,
		ClientToken: result.ClientToken,
		TTL:         result.TTL,
		Renewable:   result.Renewable,
		Message:     "login successful",
	}
}
