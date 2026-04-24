package monitor

import (
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// TransitKeyResult holds the result of a transit key check.
type TransitKeyResult struct {
	KeyName       string
	Type          string
	LatestVersion int
	Exportable    bool
	Healthy       bool
	Message       string
}

// TransitMonitor checks the health and configuration of transit keys.
type TransitMonitor struct {
	checker      *vault.TransitChecker
	logger       *log.Logger
	expectedKeys []string
}

// NewTransitMonitor creates a new TransitMonitor.
func NewTransitMonitor(checker *vault.TransitChecker, keys []string, logger *log.Logger) (*TransitMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("transit monitor: checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("transit monitor: logger must not be nil")
	}
	return &TransitMonitor{
		checker:      checker,
		logger:       logger,
		expectedKeys: keys,
	}, nil
}

// Check verifies that all expected transit keys exist and are accessible.
func (m *TransitMonitor) Check() []TransitKeyResult {
	results := make([]TransitKeyResult, 0, len(m.expectedKeys))

	for _, keyName := range m.expectedKeys {
		info, err := m.checker.GetTransitKeyInfo(keyName)
		if err != nil {
			m.logger.Printf("[transit] key %q check failed: %v", keyName, err)
			results = append(results, TransitKeyResult{
				KeyName: keyName,
				Healthy: false,
				Message: err.Error(),
			})
			continue
		}
		results = append(results, TransitKeyResult{
			KeyName:       info.Name,
			Type:          info.Type,
			LatestVersion: info.LatestVersion,
			Exportable:    info.Exportable,
			Healthy:       true,
			Message:       fmt.Sprintf("key ok, version %d", info.LatestVersion),
		})
	}

	return results
}
