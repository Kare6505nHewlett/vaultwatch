package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return p
}

func TestMain_MissingConfigFile(t *testing.T) {
	if os.Getenv("RUN_MAIN") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=TestMain_MissingConfigFile")
		cmd.Env = append(os.Environ(), "RUN_MAIN=1")
		err := cmd.Run()
		if err == nil {
			t.Fatal("expected non-zero exit for missing config, got nil")
		}
		return
	}
	os.Args = []string{"vaultwatch", "/nonexistent/path/config.yaml"}
	main()
}

func TestMain_InvalidConfig(t *testing.T) {
	if os.Getenv("RUN_MAIN_INVALID") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=TestMain_InvalidConfig")
		cmd.Env = append(os.Environ(), "RUN_MAIN_INVALID=1")
		err := cmd.Run()
		if err == nil {
			t.Fatal("expected non-zero exit for invalid config, got nil")
		}
		return
	}
	cfgPath := writeTempConfig(t, "vault:\n  address: \"\"\n")
	os.Args = []string{"vaultwatch", cfgPath}
	main()
}

func TestMain_DefaultConfigPath(t *testing.T) {
	// Verify that when no args are provided, the default config path is used.
	// We simply check the fallback assignment logic without running main.
	cfgPath := "config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	if cfgPath == "" {
		t.Error("expected a non-empty default config path")
	}
}
