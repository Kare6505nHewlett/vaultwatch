package vault

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// TokenListMonitor checks the list of active tokens and reports counts.
type TokenListMonitor struct {
	lister *TokenLister
	logger *zap.Logger
	maxTokens int
}

// TokenListResult holds the result of a token list check.
type TokenListResult struct {
	Count   int
	Warning bool
	Message string
}

// NewTokenListMonitor creates a TokenListMonitor.
// maxTokens sets the threshold above which a warning is issued (0 = no limit).
func NewTokenListMonitor(lister *TokenLister, logger *zap.Logger, maxTokens int) (*TokenListMonitor, error) {
	if lister == nil {
		return nil, fmt.Errorf("token lister must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &TokenListMonitor{
		lister:    lister,
		logger:    logger,
		maxTokens: maxTokens,
	}, nil
}

// Check lists tokens and evaluates count against the configured threshold.
func (m *TokenListMonitor) Check(ctx context.Context) (*TokenListResult, error) {
	tokens, err := m.lister.ListTokens(ctx)
	if err != nil {
		m.logger.Error("failed to list tokens", zap.Error(err))
		return nil, fmt.Errorf("list tokens: %w", err)
	}

	count := len(tokens)
	result := &TokenListResult{
		Count:   count,
		Warning: false,
		Message: fmt.Sprintf("%d active token(s) found", count),
	}

	if m.maxTokens > 0 && count > m.maxTokens {
		result.Warning = true
		result.Message = fmt.Sprintf("token count %d exceeds threshold %d", count, m.maxTokens)
		m.logger.Warn("token count exceeds threshold",
			zap.Int("count", count),
			zap.Int("max_tokens", m.maxTokens),
		)
	} else {
		m.logger.Info("token list check passed",
			zap.Int("count", count),
		)
	}

	return result, nil
}
