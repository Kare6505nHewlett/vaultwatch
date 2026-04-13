package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

const pagerDutyEventURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyNotifier sends alerts to PagerDuty via the Events API v2.
type PagerDutyNotifier struct {
	integrationKey string
	client         *http.Client
	logger         *zap.Logger
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary  string `json:"summary"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
}

// NewPagerDutyNotifier creates a new PagerDutyNotifier.
func NewPagerDutyNotifier(integrationKey string, logger *zap.Logger) (*PagerDutyNotifier, error) {
	if integrationKey == "" {
		return nil, fmt.Errorf("pagerduty: integration key must not be empty")
	}
	if logger == nil {
		return nil, fmt.Errorf("pagerduty: logger must not be nil")
	}
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		client:         &http.Client{},
		logger:         logger,
	}, nil
}

// Send dispatches an alert event to PagerDuty.
func (p *PagerDutyNotifier) Send(result CheckResult) error {
	severity := "warning"
	if result.Status == StatusExpired {
		severity = "critical"
	}

	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:  result.Message,
			Severity: severity,
			Source:   result.SecretPath,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: failed to marshal payload: %w", err)
	}

	resp, err := p.client.Post(pagerDutyEventURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status code %d", resp.StatusCode)
	}

	p.logger.Info("pagerduty alert sent",
		zap.String("secret", result.SecretPath),
		zap.String("severity", severity),
	)
	return nil
}
