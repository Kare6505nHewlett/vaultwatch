package vault

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type ReplicationStatus struct {
	Mode    string `json:"mode"`
	State   string `json:"state"`
	Healthy bool
}

type replicationResponse struct {
	Data struct {
		DR struct {
			Mode  string `json:"mode"`
			State string `json:"state"`
		} `json:"dr"`
		Performance struct {
			Mode  string `json:"mode"`
			State string `json:"state"`
		} `json:"performance"`
	} `json:"data"`
}

type ReplicationChecker struct {
	client *Client
	logger *zap.Logger
}

func NewReplicationChecker(client *Client, logger *zap.Logger) (*ReplicationChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &ReplicationChecker{client: client, logger: logger}, nil
}

func (r *ReplicationChecker) GetReplicationStatus() (*replicationResponse, error) {
	req, err := http.NewRequest(http.MethodGet, r.client.address+"/v1/sys/replication/status", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", r.client.token)

	resp, err := r.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result replicationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}
