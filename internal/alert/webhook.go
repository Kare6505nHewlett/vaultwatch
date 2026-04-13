package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// WebhookNotifier sends alert payloads to a generic HTTP webhook endpoint.
type WebhookNotifier struct {
	url    string
	client *http.Client
	logger *slog.Logger
}

type webhookPayload struct {
	Level     string    `json:"level"`
	Secret    string    `json:"secret"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message"`
}

// NewWebhookNotifier creates a new WebhookNotifier.
// Returns an error if url is empty or logger is nil.
func NewWebhookNotifier(url string, logger *slog.Logger) (*WebhookNotifier, error) {
	if url == "" {
		return nil, fmt.Errorf("webhook url must not be empty")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &WebhookNotifier{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}, nil
}

// Send dispatches an alert result to the configured webhook URL.
func (w *WebhookNotifier) Send(result monitor.CheckResult) error {
	payload := webhookPayload{
		Level:     string(result.Level),
		Secret:    result.SecretPath,
		ExpiresAt: result.ExpiresAt,
		Message:   result.Message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: failed to marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status code %d", resp.StatusCode)
	}

	w.logger.Info("webhook alert sent", "secret", result.SecretPath, "level", result.Level)
	return nil
}
