package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

func makeCheckResults() []monitor.CheckResult {
	return []monitor.CheckResult{
		{Path: "secret/db", Status: monitor.StatusOK, TTL: 48 * time.Hour},
		{Path: "secret/api", Status: monitor.StatusWarning, TTL: 2 * time.Hour},
		{Path: "secret/old", Status: monitor.StatusExpired, TTL: -1 * time.Second},
	}
}

func TestNewFormatter_DefaultsToText(t *testing.T) {
	f := NewFormatter(&bytes.Buffer{}, "")
	if f.format != FormatText {
		t.Errorf("expected default format %q, got %q", FormatText, f.format)
	}
}

func TestFormatter_WriteText_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, FormatText)
	if err := f.Write(makeCheckResults()); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	out := buf.String()
	for _, header := range []string{"PATH", "STATUS", "EXPIRES IN", "TTL"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected header %q in output", header)
		}
	}
}

func TestFormatter_WriteText_ContainsPaths(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, FormatText)
	_ = f.Write(makeCheckResults())
	out := buf.String()
	for _, path := range []string{"secret/db", "secret/api", "secret/old"} {
		if !strings.Contains(out, path) {
			t.Errorf("expected path %q in output", path)
		}
	}
}

func TestFormatter_WriteJSON_ContainsBraces(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, FormatJSON)
	if err := f.Write(makeCheckResults()); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Error("expected JSON output to start with '['")
	}
	if !strings.Contains(out, "secret/api") {
		t.Error("expected JSON to contain path 'secret/api'")
	}
}

func TestFormatDuration_Expired(t *testing.T) {
	result := formatDuration(-5 * time.Second)
	if result != "expired" {
		t.Errorf("expected 'expired', got %q", result)
	}
}

func TestFormatDuration_Hours(t *testing.T) {
	result := formatDuration(90 * time.Minute)
	if !strings.Contains(result, "h") {
		t.Errorf("expected hours in result, got %q", result)
	}
}

func TestFormatDuration_MinutesOnly(t *testing.T) {
	result := formatDuration(45 * time.Minute)
	if result != "45m" {
		t.Errorf("expected '45m', got %q", result)
	}
}
