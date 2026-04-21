package vault

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// CapabilitiesResult holds the capabilities for a given path and token.
type CapabilitiesResult struct {
	Path         string
	Capabilities []string
}

// CapabilitiesChecker checks token capabilities against Vault paths.
type CapabilitiesChecker struct {
	client *Client
	logger *zap.Logger
}

// NewCapabilitiesChecker creates a new CapabilitiesChecker.
func NewCapabilitiesChecker(client *Client, logger *zap.Logger) (*CapabilitiesChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &CapabilitiesChecker{client: client, logger: logger}, nil
}

// CheckCapabilities queries Vault for the capabilities of the current token on the given path.
func (c *CapabilitiesChecker) CheckCapabilities(ctx context.Context, path string) (*CapabilitiesResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	body := map[string]interface{}{"path": path}
	var result struct {
		Data struct {
			Capabilities []string `json:"capabilities"`
		} `json:"data"`
	}

	if err := c.client.Post(ctx, "/v1/sys/capabilities-self", body, &result); err != nil {
		c.logger.Error("failed to check capabilities", zap.String("path", path), zap.Error(err))
		return nil, fmt.Errorf("capabilities check failed for path %q: %w", path, err)
	}

	c.logger.Debug("capabilities checked", zap.String("path", path), zap.Strings("capabilities", result.Data.Capabilities))
	return &CapabilitiesResult{
		Path:         path,
		Capabilities: result.Data.Capabilities,
	}, nil
}
