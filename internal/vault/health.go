package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// HealthStatus represents the health state of a Vault instance.
type HealthStatus struct {
	Initialized bool   `json:"initialized"`
	Sealed      bool   `json:"sealed"`
	Standby     bool   `json:"standby"`
	Version     string `json:"version"`
	ClusterName string `json:"cluster_name"`
}

// HealthChecker checks the health of a Vault server.
type HealthChecker struct {
	client *Client
	logger *zap.Logger
}

// NewHealthChecker returns a new HealthChecker.
func NewHealthChecker(client *Client, logger *zap.Logger) (*HealthChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &HealthChecker{client: client, logger: logger}, nil
}

// Check queries the Vault health endpoint and returns the status.
func (h *HealthChecker) Check(ctx context.Context) (*HealthStatus, error) {
	url := h.client.address + "/v1/sys/health?standbyok=true&perfstandbyok=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building health request: %w", err)
	}

	resp, err := h.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing health request: %w", err)
	}
	defer resp.Body.Close()

	var status HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding health response: %w", err)
	}

	h.logger.Debug("vault health check",
		zap.Bool("initialized", status.Initialized),
		zap.Bool("sealed", status.Sealed),
		zap.String("version", status.Version),
	)
	return &status, nil
}
