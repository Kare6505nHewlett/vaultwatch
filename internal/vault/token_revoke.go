package vault

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// TokenRevoker handles revoking Vault tokens.
type TokenRevoker struct {
	client *Client
	logger *zap.Logger
}

// NewTokenRevoker creates a new TokenRevoker.
func NewTokenRevoker(client *Client, logger *zap.Logger) (*TokenRevoker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenRevoker{client: client, logger: logger}, nil
}

// RevokeSelf revokes the token currently used by the client.
func (r *TokenRevoker) RevokeSelf(ctx context.Context) error {
	url := r.client.Address + "/v1/auth/token/revoke-self"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", r.client.Token)

	resp, err := r.client.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status revoking self: %d", resp.StatusCode)
	}

	r.logger.Info("token revoked successfully")
	return nil
}

// RevokeAccessor revokes a token identified by its accessor.
func (r *TokenRevoker) RevokeAccessor(ctx context.Context, accessor string) error {
	if accessor == "" {
		return fmt.Errorf("accessor must not be empty")
	}

	body := fmt.Sprintf(`{"accessor":%q}`, accessor)
	url := r.client.Address + "/v1/auth/token/revoke-accessor"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url,
		stringsReader(body))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", r.client.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status revoking accessor %q: %d", accessor, resp.StatusCode)
	}

	r.logger.Info("token revoked by accessor", zap.String("accessor", accessor))
	return nil
}
