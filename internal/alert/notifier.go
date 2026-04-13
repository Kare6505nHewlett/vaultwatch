package alert

import (
	"fmt"
	"log"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarning Level = "WARNING"
	LevelExpired Level = "EXPIRED"
)

// Alert holds information about a secret expiry event.
type Alert struct {
	SecretPath string
	Level      Level
	ExpiresAt  time.Time
	TimeLeft   time.Duration
}

// Notifier defines the interface for sending alerts.
type Notifier interface {
	Send(alert Alert) error
}

// LogNotifier is a simple notifier that writes alerts to stdout/stderr.
type LogNotifier struct {
	logger *log.Logger
}

// NewLogNotifier creates a LogNotifier using the provided logger.
// If logger is nil, a default logger writing to stdout is used.
func NewLogNotifier(logger *log.Logger) *LogNotifier {
	if logger == nil {
		logger = log.Default()
	}
	return &LogNotifier{logger: logger}
}

// Send formats and logs the alert message.
func (n *LogNotifier) Send(a Alert) error {
	var msg string
	switch a.Level {
	case LevelExpired:
		msg = fmt.Sprintf(
			"[%s] Secret '%s' has EXPIRED (expired at %s)",
			a.Level,
			a.SecretPath,
			a.ExpiresAt.Format(time.RFC3339),
		)
	case LevelWarning:
		msg = fmt.Sprintf(
			"[%s] Secret '%s' expires in %s (at %s)",
			a.Level,
			a.SecretPath,
			a.TimeLeft.Round(time.Second),
			a.ExpiresAt.Format(time.RFC3339),
		)
	default:
		msg = fmt.Sprintf("[ALERT] Secret '%s': unhandled level %s", a.SecretPath, a.Level)
	}
	n.logger.Println(msg)
	return nil
}
