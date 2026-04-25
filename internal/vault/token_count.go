package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// TokenCountResult holds the result of a token count query.
type TokenCountResult struct {
	TotalCount int            `json:"total"`
	ByPolicy   map[string]int `json:"by_policy"`
}

// TokenCountChecker queries Vault for token count metrics.
type TokenCountChecker struct {
	client *Client
	logger *zap.Logger
}

// NewTokenCountChecker creates a new TokenCountChecker.
func NewTokenCountChecker(client *Client, logger *zap.Logger) (*TokenCountChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenCountChecker{client: client, logger: logger}, nil
}

// GetTokenCount returns the current token count from Vault.
func (c *TokenCountChecker) GetTokenCount() (*TokenCountResult, error) {
	req, err := http.NewRequest(http.MethodGet, c.client.Address+"/v1/auth/token/count", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.client.Token)

	resp, err := c.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var wrapper struct {
		Data TokenCountResult `json:"data"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	c.logger.Debug("token count retrieved", zap.Int("total", wrapper.Data.TotalCount))
	return &wrapper.Data, nil
}
