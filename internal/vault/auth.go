package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// TokenStatus holds metadata about the current Vault token.
type TokenStatus struct {
	Accessor   string
	Policies   []string
	TTL        time.Duration
	Renewable  bool
	ExpireTime time.Time
}

// AuthChecker checks the current token's auth status.
type AuthChecker struct {
	client *Client
	logger *zap.Logger
}

// NewAuthChecker creates a new AuthChecker.
func NewAuthChecker(client *Client, logger *zap.Logger) (*AuthChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &AuthChecker{client: client, logger: logger}, nil
}

// LookupSelf calls /v1/auth/token/lookup-self and returns the token status.
func (a *AuthChecker) LookupSelf(ctx context.Context) (*TokenStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		a.client.address+"/v1/auth/token/lookup-self", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", a.client.token)

	resp, err := a.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var payload struct {
		Data struct {
			Accessor   string   `json:"accessor"`
			Policies   []string `json:"policies"`
			TTL        int64    `json:"ttl"`
			Renewable  bool     `json:"renewable"`
			ExpireTime string   `json:"expire_time"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	status := &TokenStatus{
		Accessor:  payload.Data.Accessor,
		Policies:  payload.Data.Policies,
		TTL:       time.Duration(payload.Data.TTL) * time.Second,
		Renewable: payload.Data.Renewable,
	}
	if payload.Data.ExpireTime != "" {
		t, err := time.Parse(time.RFC3339, payload.Data.ExpireTime)
		if err == nil {
			status.ExpireTime = t
		}
	}

	a.logger.Debug("token lookup-self succeeded", zap.String("accessor", status.Accessor))
	return status, nil
}
