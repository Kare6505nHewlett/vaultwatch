package vault

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// TokenRenewer handles renewing Vault tokens via the auth/token/renew-self endpoint.
type TokenRenewer struct {
	client *Client
	logger *zap.Logger
}

// TokenRenewResult holds the result of a token renewal attempt.
type TokenRenewResult struct {
	ClientToken   string
	LeaseDuration int
	Renewable     bool
}

// NewTokenRenewer creates a new TokenRenewer.
func NewTokenRenewer(client *Client, logger *zap.Logger) (*TokenRenewer, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenRenewer{client: client, logger: logger}, nil
}

// RenewSelf attempts to renew the current token and returns the renewal result.
func (r *TokenRenewer) RenewSelf(ctx context.Context, increment int) (*TokenRenewResult, error) {
	path := "/v1/auth/token/renew-self"
	body := map[string]interface{}{}
	if increment > 0 {
		body["increment"] = increment
	}

	var resp struct {
		Auth struct {
			ClientToken   string `json:"client_token"`
			LeaseDuration int    `json:"lease_duration"`
			Renewable     bool   `json:"renewable"`
		} `json:"auth"`
	}

	if err := r.client.Post(ctx, path, body, &resp); err != nil {
		r.logger.Error("failed to renew token", zap.Error(err))
		return nil, fmt.Errorf("renew token: %w", err)
	}

	r.logger.Info("token renewed successfully",
		zap.Int("lease_duration", resp.Auth.LeaseDuration),
		zap.Bool("renewable", resp.Auth.Renewable),
	)

	return &TokenRenewResult{
		ClientToken:   resp.Auth.ClientToken,
		LeaseDuration: resp.Auth.LeaseDuration,
		Renewable:     resp.Auth.Renewable,
	}, nil
}
