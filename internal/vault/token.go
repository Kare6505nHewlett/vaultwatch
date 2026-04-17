package vault

import (
	"context"
	"fmt"
	"time"
)

// TokenInfo holds metadata about a Vault token.
type TokenInfo struct {
	Accessor   string
	TTL        time.Duration
	ExpireTime time.Time
	Renewable  bool
	Policies   []string
}

// IsExpired reports whether the token has expired.
func (t *TokenInfo) IsExpired() bool {
	return time.Now().After(t.ExpireTime)
}

// ExpiresWithin reports whether the token will expire within the given duration.
func (t *TokenInfo) ExpiresWithin(d time.Duration) bool {
	return time.Now().Add(d).After(t.ExpireTime)
}

// tokenLookupResponse mirrors the Vault API response for token self-lookup.
type tokenLookupResponse struct {
	Data struct {
		Accessor  string   `json:"accessor"`
		TTL       int      `json:"ttl"`
		Policies  []string `json:"policies"`
		Renewable bool     `json:"renewable"`
	} `json:"data"`
}

// GetTokenInfo looks up the current token's metadata via the Vault API.
func (c *Client) GetTokenInfo(ctx context.Context) (*TokenInfo, error) {
	path := "/v1/auth/token/lookup-self"

	var result tokenLookupResponse
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("token lookup failed: %w", err)
	}

	ttl := time.Duration(result.Data.TTL) * time.Second
	expireTime := time.Now().Add(ttl)

	return &TokenInfo{
		Accessor:   result.Data.Accessor,
		TTL:        ttl,
		ExpireTime: expireTime,
		Renewable:  result.Data.Renewable,
		Policies:   result.Data.Policies,
	}, nil
}
