package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// SealStatus represents the seal state of a Vault instance.
type SealStatus struct {
	Sealed      bool   `json:"sealed"`
	Initialized bool   `json:"initialized"`
	ClusterName string `json:"cluster_name"`
	Version     string `json:"version"`
}

// SealChecker checks the seal status of Vault.
type SealChecker struct {
	client *Client
	logger *zap.Logger
}

// NewSealChecker creates a new SealChecker.
func NewSealChecker(client *Client, logger *zap.Logger) (*SealChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &SealChecker{client: client, logger: logger}, nil
}

// GetSealStatus returns the current seal status from Vault.
func (s *SealChecker) GetSealStatus(ctx context.Context) (*SealStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		s.client.address+"/v1/sys/seal-status", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := s.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var status SealStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	s.logger.Debug("seal status retrieved", zap.Bool("sealed", status.Sealed))
	return &status, nil
}
