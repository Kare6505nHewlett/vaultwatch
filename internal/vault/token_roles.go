package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// TokenRole holds metadata about a Vault token role.
type TokenRole struct {
	Name            string   `json:"name"`
	AllowedPolicies []string `json:"allowed_policies"`
	Orphan          bool     `json:"orphan"`
	Renewable       bool     `json:"renewable"`
	MaxTTL          int      `json:"token_max_ttl"`
	ExplicitMaxTTL  int      `json:"explicit_max_ttl"`
}

// TokenRolesChecker fetches token role information from Vault.
type TokenRolesChecker struct {
	client *Client
	logger *log.Logger
}

// NewTokenRolesChecker creates a new TokenRolesChecker.
func NewTokenRolesChecker(client *Client, logger *log.Logger) (*TokenRolesChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenRolesChecker{client: client, logger: logger}, nil
}

// GetTokenRole retrieves details for a named token role.
func (c *TokenRolesChecker) GetTokenRole(roleName string) (*TokenRole, error) {
	if roleName == "" {
		return nil, fmt.Errorf("role name must not be empty")
	}

	path := fmt.Sprintf("/v1/auth/token/roles/%s", roleName)
	req, err := http.NewRequest(http.MethodGet, c.client.Address+path, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.client.Token)

	resp, err := c.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("token role %q not found", roleName)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for role %q", resp.StatusCode, roleName)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var wrapper struct {
		Data TokenRole `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	wrapper.Data.Name = roleName
	return &wrapper.Data, nil
}
