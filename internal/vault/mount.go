package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// MountInfo holds basic metadata about a Vault mount.
type MountInfo struct {
	Path        string
	Type        string
	Description string
	Accessor    string
}

// MountChecker lists enabled secret engine mounts from Vault.
type MountChecker struct {
	client *Client
	logger *log.Logger
}

// NewMountChecker creates a MountChecker. Returns an error if dependencies are nil.
func NewMountChecker(client *Client, logger *log.Logger) (*MountChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &MountChecker{client: client, logger: logger}, nil
}

// ListMounts returns all mounted secret engines.
func (m *MountChecker) ListMounts(ctx context.Context) ([]MountInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		m.client.address+"/v1/sys/mounts", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", m.client.token)

	resp, err := m.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var raw map[string]struct {
		Type        string `json:"type"`
		Description string `json:"description"`
		Accessor    string `json:"accessor"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	mounts := make([]MountInfo, 0, len(raw))
	for path, info := range raw {
		mounts = append(mounts, MountInfo{
			Path:        path,
			Type:        info.Type,
			Description: info.Description,
			Accessor:    info.Accessor,
		})
	}
	m.logger.Printf("listed %d mounts", len(mounts))
	return mounts, nil
}
