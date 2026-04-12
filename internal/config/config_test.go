package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	content := `
vault:
  address: "https://vault.example.com"
  token: "s.testtoken"
monitor:
  interval: 10m
  paths:
    - secret/myapp
alerts:
  warn_before_expiry: 48h
  critical_before_expiry: 2h
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Vault.Address != "https://vault.example.com" {
		t.Errorf("unexpected address: %s", cfg.Vault.Address)
	}
	if cfg.Monitor.Interval != 10*time.Minute {
		t.Errorf("unexpected interval: %v", cfg.Monitor.Interval)
	}
	if cfg.Alerts.WarnBeforeExpiry != 48*time.Hour {
		t.Errorf("unexpected warn_before_expiry: %v", cfg.Alerts.WarnBeforeExpiry)
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "s.envtoken")
	content := `vault: {}`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Vault.Address != "http://127.0.0.1:8200" {
		t.Errorf("expected default address, got: %s", cfg.Vault.Address)
	}
	if cfg.Monitor.Interval != 5*time.Minute {
		t.Errorf("expected default interval, got: %v", cfg.Monitor.Interval)
	}
	if cfg.Alerts.CriticalBeforeExpiry != time.Hour {
		t.Errorf("expected default critical expiry, got: %v", cfg.Alerts.CriticalBeforeExpiry)
	}
}

func TestLoad_MissingToken(t *testing.T) {
	os.Unsetenv("VAULT_TOKEN")
	content := `vault:
  address: "https://vault.example.com"`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
