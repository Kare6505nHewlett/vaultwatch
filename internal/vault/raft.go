package vault

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// RaftStatus holds the Raft cluster status returned by Vault.
type RaftStatus struct {
	Leader        string      `json:"leader"`
	ApplyIndex    uint64      `json:"apply_index"`
	CommitIndex   uint64      `json:"commit_index"`
	Servers       []RaftPeer  `json:"servers"`
}

// RaftPeer represents a single node in the Raft cluster.
type RaftPeer struct {
	NodeID   string `json:"node_id"`
	Address  string `json:"address"`
	Leader   bool   `json:"leader"`
	Voter    bool   `json:"voter"`
	Protocol string `json:"protocol_version"`
}

// RaftChecker queries Vault for integrated Raft storage status.
type RaftChecker struct {
	client *Client
	logger *zap.Logger
}

// NewRaftChecker creates a new RaftChecker.
func NewRaftChecker(client *Client, logger *zap.Logger) (*RaftChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &RaftChecker{client: client, logger: logger}, nil
}

// GetRaftStatus retrieves the current Raft cluster configuration from Vault.
func (r *RaftChecker) GetRaftStatus() (*RaftStatus, error) {
	url := r.client.Address + "/v1/sys/storage/raft/configuration"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building raft status request: %w", err)
	}
	req.Header.Set("X-Vault-Token", r.client.Token)

	resp, err := r.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing raft status request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from raft endpoint: %d", resp.StatusCode)
	}

	var wrapper struct {
		Data RaftStatus `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("decoding raft status response: %w", err)
	}

	r.logger.Debug("raft status retrieved", zap.String("leader", wrapper.Data.Leader))
	return &wrapper.Data, nil
}
