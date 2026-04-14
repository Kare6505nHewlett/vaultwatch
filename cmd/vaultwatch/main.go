package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/vaultwatch/internal/alert"
	"github.com/example/vaultwatch/internal/config"
	"github.com/example/vaultwatch/internal/monitor"
	"github.com/example/vaultwatch/internal/scheduler"
	"github.com/example/vaultwatch/internal/vault"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Error("failed to load config", "path", cfgPath, "error", err)
		os.Exit(1)
	}

	vaultClient, err := vault.NewClient(cfg.Vault.Address, cfg.Vault.Token)
	if err != nil {
		logger.Error("failed to create vault client", "address", cfg.Vault.Address, "error", err)
		os.Exit(1)
	}

	checker := monitor.NewChecker(vaultClient, logger)

	notifiers := buildNotifiers(cfg, logger)
	notifier := alert.NewMultiNotifier(notifiers, logger)

	jobs := scheduler.JobsFromConfig(cfg, checker, notifier, logger)
	sched := scheduler.New(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("vaultwatch started", "secrets", len(cfg.Secrets))
	if err := sched.Run(ctx, jobs); err != nil {
		logger.Error("scheduler exited with error", "error", err)
		os.Exit(1)
	}
	logger.Info("vaultwatch stopped")
}

// buildNotifiers constructs the list of alert notifiers based on the provided
// configuration. A log notifier is always included, with Slack and email
// notifiers added when their respective configuration is present.
func buildNotifiers(cfg *config.Config, logger *slog.Logger) []alert.Notifier {
	notifiers := []alert.Notifier{alert.NewLogNotifier(logger)}

	if cfg.Alerts.Slack.WebhookURL != "" {
		slackNotifier, err := alert.NewSlackNotifier(cfg.Alerts.Slack.WebhookURL, logger)
		if err != nil {
			logger.Warn("failed to create slack notifier", "error", err)
		} else {
			notifiers = append(notifiers, slackNotifier)
		}
	}

	if cfg.Alerts.Email.Host != "" {
		emailNotifier, err := alert.NewEmailNotifier(cfg.Alerts.Email, logger)
		if err != nil {
			logger.Warn("failed to create email notifier", "error", err)
		} else {
			notifiers = append(notifiers, emailNotifier)
		}
	}

	return notifiers
}
