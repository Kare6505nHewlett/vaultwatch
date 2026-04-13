package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// SlackNotifier sends alerts to a Slack webhook URL.
type SlackNotifier struct {
	webhookURL string
	logger     *zap.Logger
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackNotifier creates a new SlackNotifier with the given webhook URL and logger.
func NewSlackNotifier(webhookURL string, logger *zap.Logger) (*SlackNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack webhook URL must not be empty")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &SlackNotifier{
		webhookURL: webhookURL,
		logger:     logger,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers the alert message to the configured Slack channel.
func (s *SlackNotifier) Send(alert Alert) error {
	msg := fmt.Sprintf("[%s] Secret *%s* — %s (expires: %s)",
		alert.Level,
		alert.SecretPath,
		alert.Message,
		alert.ExpiresAt.Format(time.RFC3339),
	)

	payload, err := json.Marshal(slackPayload{Text: msg})
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send slack alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned non-2xx status: %d", resp.StatusCode)
	}

	s.logger.Info("slack alert sent",
		zap.String("secret", alert.SecretPath),
		zap.String("level", string(alert.Level)),
	)
	return nil
}
