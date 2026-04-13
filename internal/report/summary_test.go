package report_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/report"
)

func makeResults() []monitor.CheckResult {
	return []monitor.CheckResult{
		{Path: "secret/db", LeaseTTL: 72 * time.Hour, Warning: false, Expired: false},
		{Path: "secret/api", LeaseTTL: 12 * time.Hour, Warning: true, Expired: false},
		{Path: "secret/old", LeaseTTL: 0, Warning: false, Expired: true},
	}
}

func TestNewSummary_ReportCount(t *testing.T) {
	results := makeResults()
	s := report.NewSummary(results)
	if len(s.Reports) != len(results) {
		t.Fatalf("expected %d reports, got %d", len(results), len(s.Reports))
	}
}

func TestNewSummary_StatusMapping(t *testing.T) {
	s := report.NewSummary(makeResults())
	cases := []struct {
		path   string
		wantStatus report.Status
	}{
		{"secret/db", report.StatusOK},
		{"secret/api", report.StatusWarning},
		{"secret/old", report.StatusExpired},
	}
	for _, tc := range cases {
		for _, r := range s.Reports {
			if r.Path == tc.path && r.Status != tc.wantStatus {
				t.Errorf("path %s: expected status %s, got %s", tc.path, tc.wantStatus, r.Status)
			}
		}
	}
}

func TestSummary_Render_ContainsHeaders(t *testing.T) {
	s := report.NewSummary(makeResults())
	var buf bytes.Buffer
	s.Render(&buf)
	out := buf.String()
	for _, header := range []string{"PATH", "STATUS", "EXPIRES AT", "TTL"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected header %q in output", header)
		}
	}
}

func TestSummary_Render_ContainsPaths(t *testing.T) {
	s := report.NewSummary(makeResults())
	var buf bytes.Buffer
	s.Render(&buf)
	out := buf.String()
	for _, path := range []string{"secret/db", "secret/api", "secret/old"} {
		if !strings.Contains(out, path) {
			t.Errorf("expected path %q in rendered output", path)
		}
	}
}

func TestNewSummary_GeneratedAtSet(t *testing.T) {
	before := time.Now().UTC()
	s := report.NewSummary(nil)
	after := time.Now().UTC()
	if s.GeneratedAt.Before(before) || s.GeneratedAt.After(after) {
		t.Errorf("GeneratedAt %v not within expected range [%v, %v]", s.GeneratedAt, before, after)
	}
}
