package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type EngineInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Path        string
}

type EngineChecker struct {
	client *Client
	logger *zap.Logger
}

func NewEngineChecker(client *Client, logger *zap.Logger) (*EngineChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &EngineChecker{client: client, logger: logger}, nil
}

func (e *EngineChecker) ListEngines(ctx context.Context) ([]EngineInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.client.address+"/v1/sys/mounts", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", e.client.token)

	resp, err := e.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var raw map[string]struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var engines []EngineInfo
	for path, info := range raw {
		engines = append(engines, EngineInfo{
			Path:        path,
			Type:        info.Type,
			Description: info.Description,
		})
	}
	e.logger.Debug("listed secret engines", zap.Int("count", len(engines)))
	return engines, nil
}
