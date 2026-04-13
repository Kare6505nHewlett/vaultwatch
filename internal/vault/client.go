package vault

import (
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// SecretInfo holds metadata about a Vault secret or token lease.
type SecretInfo struct {
	Path      string
	LeaseTTL  time.Duration
	ExpiresAt time.Time
	IsToken   bool
}

// Client wraps the Vault API client with vaultwatch-specific helpers.
type Client struct {
	api     *vaultapi.Client
	address string
}

// NewClient creates a new Vault client using the provided address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	api.SetToken(token)

	return &Client{
		api:     api,
		address: address,
	}, nil
}

// LookupSelf returns SecretInfo for the current token.
func (c *Client) LookupSelf() (*SecretInfo, error) {
	secret, err := c.api.Auth().Token().LookupSelf()
	if err != nil {
		return nil, fmt.Errorf("looking up self token: %w", err)
	}

	ttlRaw, ok := secret.Data["ttl"]
	if !ok {
		return nil, fmt.Errorf("ttl field missing from token lookup response")
	}

	ttlJSON, ok := ttlRaw.(json.Number)
	if !ok {
		return nil, fmt.Errorf("unexpected ttl type: %T", ttlRaw)
	}

	ttlSeconds, err := ttlJSON.Int64()
	if err != nil {
		return nil, fmt.Errorf("parsing ttl value: %w", err)
	}

	ttl := time.Duration(ttlSeconds) * time.Second

	return &SecretInfo{
		Path:      "auth/token/self",
		LeaseTTL:  ttl,
		ExpiresAt: time.Now().Add(ttl),
		IsToken:   true,
	}, nil
}

// GetSecretLease returns SecretInfo for a secret at the given path.
func (c *Client) GetSecretLease(path string) (*SecretInfo, error) {
	secret, err := c.api.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path %q", path)
	}

	ttl := time.Duration(secret.LeaseDuration) * time.Second

	return &SecretInfo{
		Path:      path,
		LeaseTTL:  ttl,
		ExpiresAt: time.Now().Add(ttl),
		IsToken:   false,
	}, nil
}
