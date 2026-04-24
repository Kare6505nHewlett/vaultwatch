package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// TransitKeyInfo holds metadata about a Vault transit encryption key.
type TransitKeyInfo struct {
	Name            string
	Type            string
	DeletionAllowed bool
	Exportable      bool
	LatestVersion   int
	MinDecryptVersion int
}

// TransitChecker retrieves transit key metadata from Vault.
type TransitChecker struct {
	client *Client
	logger *log.Logger
}

// NewTransitChecker creates a new TransitChecker.
func NewTransitChecker(client *Client, logger *log.Logger) (*TransitChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("transit checker: client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("transit checker: logger must not be nil")
	}
	return &TransitChecker{client: client, logger: logger}, nil
}

// GetTransitKeyInfo fetches metadata for the named transit key.
func (t *TransitChecker) GetTransitKeyInfo(keyName string) (*TransitKeyInfo, error) {
	if keyName == "" {
		return nil, fmt.Errorf("transit checker: key name must not be empty")
	}

	url := fmt.Sprintf("%s/v1/transit/keys/%s", t.client.Address, keyName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("transit checker: failed to create request: %w", err)
	}
	req.Header.Set("X-Vault-Token", t.client.Token)

	resp, err := t.client.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transit checker: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("transit checker: key %q not found", keyName)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("transit checker: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("transit checker: failed to read response: %w", err)
	}

	var result struct {
		Data struct {
			Type              string `json:"type"`
			DeletionAllowed   bool   `json:"deletion_allowed"`
			Exportable        bool   `json:"exportable"`
			LatestVersion     int    `json:"latest_version"`
			MinDecryptVersion int    `json:"min_decryption_version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("transit checker: failed to parse response: %w", err)
	}

	return &TransitKeyInfo{
		Name:              keyName,
		Type:              result.Data.Type,
		DeletionAllowed:   result.Data.DeletionAllowed,
		Exportable:        result.Data.Exportable,
		LatestVersion:     result.Data.LatestVersion,
		MinDecryptVersion: result.Data.MinDecryptVersion,
	}, nil
}
