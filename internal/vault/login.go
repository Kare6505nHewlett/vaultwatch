package vault

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type LoginResult struct {
	ClientToken string
	Accessor    string
	TTL         int
	Renewable   bool
}

type LoginChecker struct {
	client *Client
	logger *zap.Logger
}

func NewLoginChecker(client *Client, logger *zap.Logger) (*LoginChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &LoginChecker{client: client, logger: logger}, nil
}

func (l *LoginChecker) LoginWithAppRole(roleID, secretID string) (*LoginResult, error) {
	if roleID == "" || secretID == "" {
		return nil, fmt.Errorf("roleID and secretID must not be empty")
	}

	payload := map[string]string{"role_id": roleID, "secret_id": secretID}
	resp, err := l.client.RawPost("/v1/auth/approle/login", payload)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	var result struct {
		Auth struct {
			ClientToken string `json:"client_token"`
			Accessor    string `json:"accessor"`
			LeaseDuration int  `json:"lease_duration"`
			Renewable   bool   `json:"renewable"`
		} `json:"auth"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %w", err)
	}

	l.logger.Info("AppRole login successful", zap.String("accessor", result.Auth.Accessor))
	return &LoginResult{
		ClientToken: result.Auth.ClientToken,
		Accessor:    result.Auth.Accessor,
		TTL:         result.Auth.LeaseDuration,
		Renewable:   result.Auth.Renewable,
	}, nil
}
