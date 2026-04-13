package report

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

// Format controls the output format for reports.
type Format string

const (
	FormatText Format = "text"
	FormatJSON  Format = "json"
)

// Formatter writes expiry check results to an output stream.
type Formatter struct {
	format Format
	out    io.Writer
}

// NewFormatter creates a Formatter writing to out in the given format.
func NewFormatter(out io.Writer, format Format) *Formatter {
	if format == "" {
		format = FormatText
	}
	return &Formatter{format: format, out: out}
}

// Write renders a slice of CheckResults to the formatter's output.
func (f *Formatter) Write(results []monitor.CheckResult) error {
	switch f.format {
	case FormatJSON:
		return f.writeJSON(results)
	default:
		return f.writeText(results)
	}
}

func (f *Formatter) writeText(results []monitor.CheckResult) error {
	w := tabwriter.NewWriter(f.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tSTATUS\tEXPIRES IN\tTTL")
	fmt.Fprintln(w, strings.Repeat("-", 60))
	for _, r := range results {
		expiry := formatDuration(r.TTL)
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", r.Path, r.Status, expiry, int(r.TTL.Seconds()))
	}
	return w.Flush()
}

func (f *Formatter) writeJSON(results []monitor.CheckResult) error {
	fmt.Fprintln(f.out, "[")
	for i, r := range results {
		comma := ","
		if i == len(results)-1 {
			comma = ""
		}
		fmt.Fprintf(f.out, "  {\"path\": %q, \"status\": %q, \"ttl_seconds\": %d}%s\n",
			r.Path, r.Status, int(r.TTL.Seconds()), comma)
	}
	fmt.Fprintln(f.out, "]")
	return nil
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "expired"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
