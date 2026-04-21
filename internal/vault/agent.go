package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// AgentInfo holds metadata about a running Vault Agent instance.
type AgentInfo struct {
	Version    string    `json:"version"`
	CacheState string    `json:"cache_state"`
	CheckedAt  time.Time `json:"checked_at"`
}

// AgentChecker queries a Vault Agent API endpoint for status.
type AgentChecker struct {
	client *Client
	logger *zap.Logger
}

// NewAgentChecker returns an AgentChecker or an error if dependencies are nil.
func NewAgentChecker(client *Client, logger *zap.Logger) (*AgentChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &AgentChecker{client: client, logger: logger}, nil
}

// GetAgentInfo queries the Vault Agent /v1/agent/v1/metrics-like health endpoint.
// In practice, agents expose /v1/sys/health via proxy; we call /agent/v1/cache.
func (a *AgentChecker) GetAgentInfo() (*AgentInfo, error) {
	url := fmt.Sprintf("%s/v1/agent/v1/cache", a.client.Address)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building agent request: %w", err)
	}
	req.Header.Set("X-Vault-Token", a.client.Token)

	resp, err := a.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("agent request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from agent: %d", resp.StatusCode)
	}

	var payload struct {
		Data struct {
			CacheState string `json:"cache_state"`
			Version    string `json:"version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decoding agent response: %w", err)
	}

	info := &AgentInfo{
		Version:    payload.Data.Version,
		CacheState: payload.Data.CacheState,
		CheckedAt:  time.Now().UTC(),
	}

	a.logger.Info("agent info retrieved",
		zap.String("version", info.Version),
		zap.String("cache_state", info.CacheState),
	)
	return info, nil
}
