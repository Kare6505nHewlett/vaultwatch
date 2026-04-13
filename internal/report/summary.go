package report

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// Status represents the expiry status of a secret.
type Status string

const (
	StatusOK      Status = "OK"
	StatusWarning Status = "WARNING"
	StatusExpired Status = "EXPIRED"
)

// SecretReport holds the summary information for a single secret.
type SecretReport struct {
	Path      string
	LeaseTTL  time.Duration
	Status    Status
	ExpiresAt time.Time
}

// Summary aggregates reports for all monitored secrets.
type Summary struct {
	GeneratedAt time.Time
	Reports     []SecretReport
}

// NewSummary builds a Summary from a slice of monitor.CheckResult.
func NewSummary(results []monitor.CheckResult) Summary {
	s := Summary{
		GeneratedAt: time.Now().UTC(),
		Reports:     make([]SecretReport, 0, len(results)),
	}
	for _, r := range results {
		s.Reports = append(s.Reports, SecretReport{
			Path:      r.Path,
			LeaseTTL:  r.LeaseTTL,
			Status:    toStatus(r),
			ExpiresAt: time.Now().UTC().Add(r.LeaseTTL),
		})
	}
	return s
}

// Render writes a human-readable table of the summary to w.
func (s Summary) Render(w io.Writer) {
	fmt.Fprintf(w, "VaultWatch Report — %s\n\n", s.GeneratedAt.Format(time.RFC3339))
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tSTATUS\tEXPIRES AT\tTTL")
	fmt.Fprintln(tw, "----\t------\t----------\t---")
	for _, r := range s.Reports {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			r.Path,
			r.Status,
			r.ExpiresAt.Format(time.RFC3339),
			r.LeaseTTL.Round(time.Second),
		)
	}
	tw.Flush()
}

func toStatus(r monitor.CheckResult) Status {
	switch {
	case r.Expired:
		return StatusExpired
	case r.Warning:
		return StatusWarning
	default:
		return StatusOK
	}
}
