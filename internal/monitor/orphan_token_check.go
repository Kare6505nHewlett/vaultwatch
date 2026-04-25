package monitor

import (
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// OrphanTokenMonitor checks whether monitored tokens are orphans.
type OrphanTokenMonitor struct {
	checker *vault.OrphanTokenChecker
	tokens  []string
	logger  *log.Logger
}

// OrphanTokenResult holds the result of an orphan token check.
type OrphanTokenResult struct {
	Token  string
	Orphan bool
	TTL    int
	Err    error
}

// NewOrphanTokenMonitor returns a new OrphanTokenMonitor.
func NewOrphanTokenMonitor(checker *vault.OrphanTokenChecker, tokens []string, logger *log.Logger) (*OrphanTokenMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("orphan token checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &OrphanTokenMonitor{
		checker: checker,
		tokens:  tokens,
		logger:  logger,
	}, nil
}

// CheckAll runs orphan checks against all configured tokens.
func (m *OrphanTokenMonitor) CheckAll() []OrphanTokenResult {
	results := make([]OrphanTokenResult, 0, len(m.tokens))
	for _, tok := range m.tokens {
		info, err := m.checker.IsOrphanToken(tok)
		if err != nil {
			m.logger.Printf("[orphan-monitor] error checking token %s: %v", tok, err)
			results = append(results, OrphanTokenResult{Token: tok, Err: err})
			continue
		}
		if !info.Orphan {
			m.logger.Printf("[orphan-monitor] WARNING: token %s is not an orphan (ttl=%d)", tok, info.TTL)
		} else {
			m.logger.Printf("[orphan-monitor] token %s is orphan (ttl=%d)", tok, info.TTL)
		}
		results = append(results, OrphanTokenResult{
			Token:  tok,
			Orphan: info.Orphan,
			TTL:    info.TTL,
		})
	}
	return results
}
