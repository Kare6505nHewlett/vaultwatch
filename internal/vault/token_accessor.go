package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// TokenAccessorInfo holds metadata returned when looking up a token by its accessor.
type TokenAccessorInfo struct {
	Accessor    string
	DisplayName string
	Policies    []string
	TTL         time.Duration
	ExpireTime  time.Time
	Orphan      bool
	Renewable   bool
}

// TokenAccessorChecker looks up Vault token details using a token accessor.
type TokenAccessorChecker struct {
	client *Client
	logger *log.Logger
}

// NewTokenAccessorChecker creates a new TokenAccessorChecker.
// Returns an error if client or logger is nil.
func NewTokenAccessorChecker(client *Client, logger *log.Logger) (*TokenAccessorChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenAccessorChecker{client: client, logger: logger}, nil
}

// LookupByAccessor queries Vault for token information using the given accessor string.
// Requires a token with the 'auth/token/lookup-accessor' capability.
func (c *TokenAccessorChecker) LookupByAccessor(accessor string) (*TokenAccessorInfo, error) {
	if strings.TrimSpace(accessor) == "" {
		return nil, fmt.Errorf("accessor must not be empty")
	}

	body, err := json.Marshal(map[string]string{"accessor": accessor})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/v1/auth/token/lookup-accessor", c.client.Address)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.client.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("accessor not found: %s", accessor)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from Vault", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Data struct {
			Accessor    string   `json:"accessor"`
			DisplayName string   `json:"display_name"`
			Policies    []string `json:"policies"`
			TTL         int      `json:"ttl"`
			ExpireTime  string   `json:"expire_time"`
			Orphan      bool     `json:"orphan"`
			Renewable   bool     `json:"renewable"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	info := &TokenAccessorInfo{
		Accessor:    result.Data.Accessor,
		DisplayName: result.Data.DisplayName,
		Policies:    result.Data.Policies,
		TTL:         time.Duration(result.Data.TTL) * time.Second,
		Orphan:      result.Data.Orphan,
		Renewable:   result.Data.Renewable,
	}

	if result.Data.ExpireTime != "" {
		parsed, err := time.Parse(time.RFC3339, result.Data.ExpireTime)
		if err != nil {
			c.logger.Printf("[WARN] could not parse expire_time %q: %v", result.Data.ExpireTime, err)
		} else {
			info.ExpireTime = parsed
		}
	}

	c.logger.Printf("[DEBUG] looked up accessor %s: display_name=%s ttl=%s",
		accessor, info.DisplayName, info.TTL)

	return info, nil
}
