package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type NamespaceInfo struct {
	Path string
	ID   string
}

type NamespaceChecker struct {
	client *Client
	logger *zap.Logger
}

func NewNamespaceChecker(client *Client, logger *zap.Logger) (*NamespaceChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &NamespaceChecker{client: client, logger: logger}, nil
}

func (n *NamespaceChecker) ListNamespaces(ctx context.Context) ([]NamespaceInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, n.client.address+"/v1/sys/namespaces", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", n.client.token)

	resp, err := n.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	var result struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	var namespaces []NamespaceInfo
	for _, k := range result.Data.Keys {
		namespaces = append(namespaces, NamespaceInfo{Path: k, ID: k})
	}

	n.logger.Info("listed namespaces", zap.Int("count", len(namespaces)))
	return namespaces, nil
}
