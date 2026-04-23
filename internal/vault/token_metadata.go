package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// TokenMetadata holds metadata about a Vault token.
type TokenMetadata struct {
	DisplayName string            `json:"display_name"`
	Policies    []string          `json:"policies"`
	Meta        map[string]string `json:"meta"`
	EntityID    string            `json:"entity_id"`
	Orphan      bool              `json:"orphan"`
}

// TokenMetadataChecker retrieves metadata for a Vault token.
type TokenMetadataChecker struct {
	client *Client
	logger *zap.Logger
}

// NewTokenMetadataChecker creates a new TokenMetadataChecker.
func NewTokenMetadataChecker(client *Client, logger *zap.Logger) (*TokenMetadataChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenMetadataChecker{client: client, logger: logger}, nil
}

// GetTokenMetadata fetches metadata for the given token accessor.
func (c *TokenMetadataChecker) GetTokenMetadata(accessor string) (*TokenMetadata, error) {
	if accessor == "" {
		return nil, fmt.Errorf("accessor must not be empty")
	}

	url := fmt.Sprintf("%s/v1/auth/token/lookup-accessor", c.client.Address)
	body := fmt.Sprintf(`{"accessor":%q}`, accessor)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.client.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("token accessor not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data TokenMetadata `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.Debug("fetched token metadata", zap.String("accessor", accessor))
	return &result.Data, nil
}
