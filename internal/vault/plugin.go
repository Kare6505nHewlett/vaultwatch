package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// PluginInfo holds metadata about a registered Vault plugin.
type PluginInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Builtin bool   `json:"builtin"`
}

// PluginChecker checks registered plugins in Vault.
type PluginChecker struct {
	client *Client
	logger *zap.Logger
}

// NewPluginChecker returns a new PluginChecker or an error if dependencies are nil.
func NewPluginChecker(client *Client, logger *zap.Logger) (*PluginChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &PluginChecker{client: client, logger: logger}, nil
}

// ListPlugins returns all registered plugins of the given type (e.g. "secret", "auth", "database").
func (p *PluginChecker) ListPlugins(ctx context.Context, pluginType string) ([]PluginInfo, error) {
	url := fmt.Sprintf("%s/v1/sys/plugins/catalog/%s", p.client.address, pluginType)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", p.client.token)

	resp, err := p.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing plugins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d listing plugins", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding plugin list: %w", err)
	}

	plugins := make([]PluginInfo, 0, len(result.Data.Keys))
	for _, name := range result.Data.Keys {
		plugins = append(plugins, PluginInfo{
			Name: name,
			Type: pluginType,
		})
	}

	p.logger.Info("listed plugins", zap.String("type", pluginType), zap.Int("count", len(plugins)))
	return plugins, nil
}
