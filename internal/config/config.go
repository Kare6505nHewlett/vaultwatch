package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all vaultwatch configuration.
type Config struct {
	Vault    VaultConfig    `yaml:"vault"`
	Monitor  MonitorConfig  `yaml:"monitor"`
	Alert    AlertConfig    `yaml:"alert"`
}

// VaultConfig holds Vault connection settings.
type VaultConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
	Secrets []string `yaml:"secrets"`
}

// MonitorConfig holds monitoring schedule settings.
type MonitorConfig struct {
	Interval    time.Duration `yaml:"interval"`
	WarnBefore  time.Duration `yaml:"warn_before"`
}

// AlertConfig holds alert delivery configuration.
type AlertConfig struct {
	Log   LogAlertConfig   `yaml:"log"`
	Slack SlackAlertConfig `yaml:"slack"`
	Email EmailAlertConfig `yaml:"email"`
}

// LogAlertConfig configures log-based alerts.
type LogAlertConfig struct {
	Enabled bool `yaml:"enabled"`
}

// SlackAlertConfig configures Slack webhook alerts.
type SlackAlertConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
}

// EmailAlertConfig configures SMTP email alerts.
type EmailAlertConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
	Password string   `yaml:"password"`
}

// Load reads and parses the config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := applyDefaults(&cfg); err != nil {
		return nil, err
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyDefaults(cfg *Config) error {
	if cfg.Vault.Address == "" {
		cfg.Vault.Address = "http://127.0.0.1:8200"
	}
	if cfg.Monitor.Interval == 0 {
		cfg.Monitor.Interval = 5 * time.Minute
	}
	if cfg.Monitor.WarnBefore == 0 {
		cfg.Monitor.WarnBefore = 72 * time.Hour
	}
	if cfg.Alert.Email.Port == 0 {
		cfg.Alert.Email.Port = 587
	}
	return nil
}

func validate(cfg *Config) error {
	if cfg.Vault.Token == "" {
		return errors.New("vault.token is required")
	}
	return nil
}
