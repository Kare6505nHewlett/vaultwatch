package vault

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// AuditDevice represents a Vault audit device.
type AuditDevice struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Enabled     bool
}

// AuditChecker checks the enabled audit devices on a Vault server.
type AuditChecker struct {
	client *Client
	logger *zap.Logger
}

// NewAuditChecker returns a new AuditChecker or an error if dependencies are nil.
func NewAuditChecker(client *Client, logger *zap.Logger) (*AuditChecker, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &AuditChecker{client: client, logger: logger}, nil
}

// ListAuditDevices returns the audit devices enabled on Vault.
func (a *AuditChecker) ListAuditDevices() ([]AuditDevice, error) {
	req, err := http.NewRequest(http.MethodGet, a.client.address+"/v1/sys/audit", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", a.client.token)

	resp, err := a.client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var raw map[string]struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	devices := make([]AuditDevice, 0, len(raw))
	for path, d := range raw {
		devices = append(devices, AuditDevice{
			Path:        path,
			Type:        d.Type,
			Description: d.Description,
			Enabled:     true,
		})
	}
	a.logger.Info("listed audit devices", zap.Int("count", len(devices)))
	return devices, nil
}
