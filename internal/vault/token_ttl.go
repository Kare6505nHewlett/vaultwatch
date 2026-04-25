package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// TokenTTLInfo holds TTL information for a Vault token.
type TokenTTLInfo struct {
	TTL         time.Duration
	CreationTTL time.Duration
	ExpireTime  *time.Time
	Renewable   bool
}

// TokenTTLChecker fetches TTL details for the current token.
type TokenTTLChecker struct {
	client *Client
	logger *zap.Logger
}

// NewTokenTTLChecker creates a new TokenTTLChecker.
func NewTokenTTLChecker(client *Client, logger *zap.Logger) (*TokenTTLChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenTTLChecker{client: client, logger: logger}, nil
}

// GetTokenTTL calls the Vault token lookup-self endpoint and returns TTL info.
func (t *TokenTTLChecker) GetTokenTTL() (*TokenTTLInfo, error) {
	req, err := http.NewRequest(http.MethodGet, t.client.Address+"/v1/auth/token/lookup-self", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", t.client.Token)

	resp, err := t.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			TTL         int    `json:"ttl"`
			CreationTTL int    `json:"creation_ttl"`
			ExpireTime  string `json:"expire_time"`
			Renewable   bool   `json:"renewable"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	info := &TokenTTLInfo{
		TTL:         time.Duration(result.Data.TTL) * time.Second,
		CreationTTL: time.Duration(result.Data.CreationTTL) * time.Second,
		Renewable:   result.Data.Renewable,
	}

	if result.Data.ExpireTime != "" {
		parsed, err := time.Parse(time.RFC3339, result.Data.ExpireTime)
		if err == nil {
			info.ExpireTime = &parsed
		} else {
			t.logger.Warn("failed to parse expire_time", zap.String("value", result.Data.ExpireTime))
		}
	}

	t.logger.Debug("token TTL fetched", zap.Duration("ttl", info.TTL), zap.Bool("renewable", info.Renewable))
	return info, nil
}
