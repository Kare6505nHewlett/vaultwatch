package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// SecretMetadata holds metadata for a KV secret path.
type SecretMetadata struct {
	Path        string
	Version     int
	CreatedTime time.Time
	DeletedTime *time.Time
	Destroyed   bool
}

// LookupSecret retrieves metadata for a KV v2 secret at the given mount and path.
func (c *Client) LookupSecret(ctx context.Context, mount, path string) (*SecretMetadata, error) {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", c.address, mount, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s/%s", mount, path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s/%s", resp.StatusCode, mount, path)
	}

	var body struct {
		Data struct {
			Versions map[string]struct {
				CreatedTime  time.Time  `json:"created_time"`
				DeletionTime *time.Time `json:"deletion_time"`
				Destroyed    bool       `json:"destroyed"`
			} `json:"versions"`
			CurrentVersion int `json:"current_version"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	version := body.Data.CurrentVersion
	versionKey := fmt.Sprintf("%d", version)
	v, ok := body.Data.Versions[versionKey]
	if !ok {
		return nil, fmt.Errorf("version %d not found in metadata", version)
	}

	c.logger.Debug("looked up secret metadata",
		zap.String("mount", mount),
		zap.String("path", path),
		zap.Int("version", version),
	)

	return &SecretMetadata{
		Path:        fmt.Sprintf("%s/%s", mount, path),
		Version:     version,
		CreatedTime: v.CreatedTime,
		DeletedTime: v.DeletionTime,
		Destroyed:   v.Destroyed,
	}, nil
}
