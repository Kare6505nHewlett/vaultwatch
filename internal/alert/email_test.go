package alert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

func newTestEmailConfig() EmailConfig {
	return EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		From:     "alerts@example.com",
		To:       []string{"admin@example.com"},
		Password: "secret",
	}
}

func TestNewEmailNotifier_Valid(t *testing.T) {
	logger := zap.NewNop()
	cfg := newTestEmailConfig()

	n, err := NewEmailNotifier(cfg, logger)

	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.Equal(t, "smtp.example.com", n.host)
	assert.Equal(t, 587, n.port)
}

func TestNewEmailNotifier_DefaultPort(t *testing.T) {
	cfg := newTestEmailConfig()
	cfg.Port = 0

	n, err := NewEmailNotifier(cfg, nil)

	assert.NoError(t, err)
	assert.Equal(t, 587, n.port)
}

func TestNewEmailNotifier_NilLogger(t *testing.T) {
	cfg := newTestEmailConfig()

	n, err := NewEmailNotifier(cfg, nil)

	assert.NoError(t, err)
	assert.NotNil(t, n.logger)
}

func TestNewEmailNotifier_EmptyHost(t *testing.T) {
	cfg := newTestEmailConfig()
	cfg.Host = ""

	_, err := NewEmailNotifier(cfg, zap.NewNop())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")
}

func TestNewEmailNotifier_EmptyFrom(t *testing.T) {
	cfg := newTestEmailConfig()
	cfg.From = ""

	_, err := NewEmailNotifier(cfg, zap.NewNop())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from")
}

func TestNewEmailNotifier_NoRecipients(t *testing.T) {
	cfg := newTestEmailConfig()
	cfg.To = []string{}

	_, err := NewEmailNotifier(cfg, zap.NewNop())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipient")
}

func TestEmailNotifier_Send_FailsWithBadHost(t *testing.T) {
	cfg := newTestEmailConfig()
	cfg.Host = "localhost"
	cfg.Port = 19999 // unlikely to be listening

	n, err := NewEmailNotifier(cfg, zap.NewNop())
	assert.NoError(t, err)

	result := monitor.CheckResult{
		SecretPath: "secret/db/password",
		Status:     monitor.StatusWarning,
		TTL:        48 * time.Hour,
		Message:    "expires soon",
	}

	err = n.Send(result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email send failed")
}
