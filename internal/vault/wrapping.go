package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// WrappingInfo holds metadata about a wrapped token response.
type WrappingInfo struct {
	Token          string
	Accessor       string
	TTL            time.Duration
	CreationTime   time.Time
	CreationPath   string
}

// WrappingChecker checks wrapping token lookup via Vault.
type WrappingChecker struct {
	client *Client
	logger *zap.Logger
}

// NewWrappingChecker creates a new WrappingChecker.
func NewWrappingChecker(client *Client, logger *zap.Logger) (*WrappingChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &WrappingChecker{client: client, logger: logger}, nil
}

// LookupWrappingToken looks up metadata for a wrapping token.
func (w *WrappingChecker) LookupWrappingToken(token string) (*WrappingInfo, error) {
	if token == "" {
		return nil, fmt.Errorf("wrapping token must not be empty")
	}

	req, err := http.NewRequest(http.MethodPost, w.client.Address+"/v1/sys/wrapping/lookup", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := w.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Token        string `json:"token"`
			Accessor     string `json:"accessor"`
			TTL          int    `json:"ttl"`
			CreationTime string `json:"creation_time"`
			CreationPath string `json:"creation_path"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	ct, _ := time.Parse(time.RFC3339, result.Data.CreationTime)

	w.logger.Info("wrapping token looked up", zap.String("accessor", result.Data.Accessor))

	return &WrappingInfo{
		Token:        result.Data.Token,
		Accessor:     result.Data.Accessor,
		TTL:          time.Duration(result.Data.TTL) * time.Second,
		CreationTime: ct,
		CreationPath: result.Data.CreationPath,
	}, nil
}
