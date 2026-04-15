package vault

import (
	"context"
	"fmt"
	"log/slog"
)

// Rotator handles rotation of Vault secrets.
type Rotator struct {
	client *Client
	logger *slog.Logger
}

// RotateResult holds the outcome of a rotation attempt.
type RotateResult struct {
	Path    string
	Success bool
	Message string
}

// NewRotator creates a new Rotator. Returns an error if client or logger is nil.
func NewRotator(client *Client, logger *slog.Logger) (*Rotator, error) {
	if client == nil {
		return nil, fmt.Errorf("vault client must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &Rotator{client: client, logger: logger}, nil
}

// RotateSecret triggers a credential rotation for the given mount path.
// It calls the Vault sys/rotate endpoint for the provided mount.
func (r *Rotator) RotateSecret(ctx context.Context, mountPath string) (*RotateResult, error) {
	if mountPath == "" {
		return nil, fmt.Errorf("mountPath must not be empty")
	}

	r.logger.Info("rotating secret", "mount", mountPath)

	path := fmt.Sprintf("/v1/%s/rotate", mountPath)
	resp, err := r.client.RawPost(ctx, path, nil)
	if err != nil {
		return &RotateResult{Path: mountPath, Success: false, Message: err.Error()}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := fmt.Sprintf("unexpected status %d from Vault", resp.StatusCode)
		r.logger.Warn("rotation failed", "mount", mountPath, "status", resp.StatusCode)
		return &RotateResult{Path: mountPath, Success: false, Message: msg}, fmt.Errorf(msg)
	}

	r.logger.Info("rotation succeeded", "mount", mountPath)
	return &RotateResult{Path: mountPath, Success: true, Message: "rotated successfully"}, nil
}
