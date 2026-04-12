package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for vaultwatch.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Alerts  AlertsConfig  `yaml:"alerts"`
	Monitor MonitorConfig `yaml:"monitor"`
}

// VaultConfig contains Vault connection settings.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	Namespace string `yaml:"namespace"`
}

// AlertsConfig defines alerting thresholds and channels.
type AlertsConfig struct {
	WarnBeforeExpiry  time.Duration `yaml:"warn_before_expiry"`
	CriticalBeforeExpiry time.Duration `yaml:"critical_before_expiry"`
	SlackWebhookURL   string        `yaml:"slack_webhook_url"`
	EmailRecipients   []string      `yaml:"email_recipients"`
}

// MonitorConfig controls polling behaviour.
type MonitorConfig struct {
	Interval time.Duration `yaml:"interval"`
	Paths    []string      `yaml:"paths"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate checks required fields and applies defaults.
func (c *Config) validate() error {
	if c.Vault.Address == "" {
		c.Vault.Address = "http://127.0.0.1:8200"
	}
	if c.Vault.Token == "" {
		c.Vault.Token = os.Getenv("VAULT_TOKEN")
	}
	if c.Vault.Token == "" {
		return fmt.Errorf("vault token must be set via config or VAULT_TOKEN env var")
	}
	if c.Monitor.Interval == 0 {
		c.Monitor.Interval = 5 * time.Minute
	}
	if c.Alerts.WarnBeforeExpiry == 0 {
		c.Alerts.WarnBeforeExpiry = 24 * time.Hour
	}
	if c.Alerts.CriticalBeforeExpiry == 0 {
		c.Alerts.CriticalBeforeExpiry = 1 * time.Hour
	}
	return nil
}
