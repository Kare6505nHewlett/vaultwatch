package alert

import (
	"fmt"
	"net/smtp"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// EmailNotifier sends alert notifications via SMTP email.
type EmailNotifier struct {
	host     string
	port     int
	from     string
	to       []string
	password string
	logger   *zap.Logger
}

// EmailConfig holds configuration for the email notifier.
type EmailConfig struct {
	Host     string
	Port     int
	From     string
	To       []string
	Password string
}

// NewEmailNotifier creates a new EmailNotifier.
// Returns an error if required fields are missing.
func NewEmailNotifier(cfg EmailConfig, logger *zap.Logger) (*EmailNotifier, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("email host must not be empty")
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("email from address must not be empty")
	}
	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("email must have at least one recipient")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	return &EmailNotifier{
		host:     cfg.Host,
		port:     cfg.Port,
		from:     cfg.From,
		to:       cfg.To,
		password: cfg.Password,
		logger:   logger,
	}, nil
}

// Send delivers an alert notification via email.
func (e *EmailNotifier) Send(result monitor.CheckResult) error {
	subject := fmt.Sprintf("[VaultWatch] %s: %s", result.Status, result.SecretPath)
	body := fmt.Sprintf(
		"Secret: %s\nStatus: %s\nExpires In: %s\nMessage: %s",
		result.SecretPath,
		result.Status,
		result.TTL.String(),
		result.Message,
	)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.from, e.to[0], subject, body,
	))

	addr := fmt.Sprintf("%s:%d", e.host, e.port)
	var auth smtp.Auth
	if e.password != "" {
		auth = smtp.PlainAuth("", e.from, e.password, e.host)
	}

	if err := smtp.SendMail(addr, auth, e.from, e.to, msg); err != nil {
		e.logger.Error("failed to send email alert", zap.String("path", result.SecretPath), zap.Error(err))
		return fmt.Errorf("email send failed: %w", err)
	}

	e.logger.Info("email alert sent", zap.String("path", result.SecretPath), zap.String("status", string(result.Status)))
	return nil
}
