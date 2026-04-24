package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// TokenListResult holds the list of token accessor IDs returned by Vault.
type TokenListResult struct {
	Accessors []string
}

// TokenLister lists active token accessors from Vault.
type TokenLister struct {
	client *Client
	logger *zap.Logger
}

// NewTokenLister creates a new TokenLister. Returns an error if client or logger is nil.
func NewTokenLister(client *Client, logger *zap.Logger) (*TokenLister, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenLister{client: client, logger: logger}, nil
}

// ListTokens retrieves all active token accessors from Vault.
func (tl *TokenLister) ListTokens() (*TokenListResult, error) {
	req, err := http.NewRequest("LIST", tl.client.Address+"/v1/auth/token/accessors", nil)
	if err != nil {
		return nil, fmt.Errorf("building list request: %w", err)
	}
	req.Header.Set("X-Vault-Token", tl.client.Token)

	resp, err := tl.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing list request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status listing tokens: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decoding list response: %w", err)
	}

	tl.logger.Info("listed token accessors", zap.Int("count", len(result.Data.Keys)))
	return &TokenListResult{Accessors: result.Data.Keys}, nil
}
