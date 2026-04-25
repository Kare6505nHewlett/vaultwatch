package vault

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// TokenExpireInfo holds expiry details for a Vault token.
type TokenExpireInfo struct {
	TokenID    string
	ExpireTime time.Time
	TTL        time.Duration
	Renewable  bool
}

// TokenExpireChecker checks when a token will expire.
type TokenExpireChecker struct {
	client HTTPClient
	logger *zap.Logger
}

// NewTokenExpireChecker creates a new TokenExpireChecker.
func NewTokenExpireChecker(client HTTPClient, logger *zap.Logger) (*TokenExpireChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenExpireChecker{client: client, logger: logger}, nil
}

// GetTokenExpiry returns expiry information for the given token.
func (c *TokenExpireChecker) GetTokenExpiry(ctx context.Context, token string) (*TokenExpireInfo, error) {
	if token == "" {
		return nil, fmt.Errorf("token must not be empty")
	}

	path := "/v1/auth/token/lookup"
	body := map[string]string{"token": token}

	var result struct {
		Data struct {
			ID        string `json:"id"`
			TTL       int    `json:"ttl"`
			Renewable bool   `json:"renewable"`
			ExpireTime string `json:"expire_time"`
		} `json:"data"`
	}

	if err := c.client.PostJSON(ctx, path, body, &result); err != nil {
		c.logger.Error("failed to lookup token expiry", zap.String("path", path), zap.Error(err))
		return nil, fmt.Errorf("lookup token expiry: %w", err)
	}

	ttl := time.Duration(result.Data.TTL) * time.Second
	var expireTime time.Time
	if result.Data.ExpireTime != "" {
		parsed, err := time.Parse(time.RFC3339, result.Data.ExpireTime)
		if err == nil {
			expireTime = parsed
		}
	}
	if expireTime.IsZero() && ttl > 0 {
		expireTime = time.Now().Add(ttl)
	}

	c.logger.Debug("token expiry retrieved",
		zap.String("token_id", result.Data.ID),
		zap.Duration("ttl", ttl),
	)

	return &TokenExpireInfo{
		TokenID:    result.Data.ID,
		ExpireTime: expireTime,
		TTL:        ttl,
		Renewable:  result.Data.Renewable,
	}, nil
}
