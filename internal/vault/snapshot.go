package vault

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// SnapshotChecker checks if a Vault raft snapshot can be taken.
type SnapshotChecker struct {
	client *Client
	logger *zap.Logger
}

// SnapshotResult holds metadata about a snapshot attempt.
type SnapshotResult struct {
	Available bool
	Bytes     int64
	Error     string
}

// NewSnapshotChecker creates a new SnapshotChecker.
func NewSnapshotChecker(client *Client, logger *zap.Logger) (*SnapshotChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &SnapshotChecker{client: client, logger: logger}, nil
}

// CheckSnapshot attempts a HEAD request against the raft snapshot endpoint.
func (s *SnapshotChecker) CheckSnapshot(ctx context.Context) (*SnapshotResult, error) {
	url := fmt.Sprintf("%s/v1/sys/storage/raft/snapshot", s.client.Address)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building snapshot request: %w", err)
	}
	req.Header.Set("X-Vault-Token", s.client.Token)

	resp, err := s.client.HTTP.Do(req)
	if err != nil {
		return &SnapshotResult{Available: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("snapshot endpoint returned non-200", zap.Int("status", resp.StatusCode))
		return &SnapshotResult{Available: false, Error: fmt.Sprintf("status %d", resp.StatusCode)}, nil
	}

	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading snapshot body: %w", err)
	}

	s.logger.Info("snapshot available", zap.Int64("bytes", n))
	return &SnapshotResult{Available: true, Bytes: n}, nil
}
