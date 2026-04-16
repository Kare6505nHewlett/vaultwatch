package vault

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

// AuthRenewer renews the Vault auth token.
type AuthRenewer struct {
	client *Client
	logger *slog.Logger
}

// NewAuthRenewer creates a new AuthRenewer.
func NewAuthRenewer(client *Client, logger *slog.Logger) (*AuthRenewer, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &AuthRenewer{client: client, logger: logger}, nil
}

// RenewSelf renews the current token and returns the new TTL in seconds.
func (r *AuthRenewer) RenewSelf(ctx context.Context, increment int) (int, error) {
	body := map[string]any{"increment": increment}
	resp, err := r.client.RawPost(ctx, "/v1/auth/token/renew-self", body)
	if err != nil {
		return 0, fmt.Errorf("renew-self request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("renew-self returned status %d", resp.StatusCode)
	}

	var result struct {
		Auth struct {
			LeaseDuration int `json:"lease_duration"`
		} `json:"auth"`
	}
	if err := decodeJSON(resp.Body, &result); err != nil {
		return 0, fmt.Errorf("failed to decode renew-self response: %w", err)
	}

	ttl := result.Auth.LeaseDuration
	r.logger.InfoContext(ctx, "token renewed", "ttl_seconds", ttl)
	return ttl, nil
}
