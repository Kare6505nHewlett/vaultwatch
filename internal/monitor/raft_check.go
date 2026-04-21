package monitor

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// RaftMonitor checks the health of the Vault Raft cluster.
type RaftMonitor struct {
	checker *vault.RaftChecker
	logger  *zap.Logger
}

// RaftResult holds the outcome of a Raft cluster health check.
type RaftResult struct {
	Leader      string
	TotalPeers  int
	VoterCount  int
	Healthy     bool
	Message     string
}

// NewRaftMonitor creates a new RaftMonitor.
func NewRaftMonitor(checker *vault.RaftChecker, logger *zap.Logger) (*RaftMonitor, error) {
	if checker == nil {
		return nil, fmt.Errorf("raft checker must not be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger must not be nil")
	}
	return &RaftMonitor{checker: checker, logger: logger}, nil
}

// Check retrieves Raft status and evaluates cluster health.
func (m *RaftMonitor) Check() (*RaftResult, error) {
	status, err := m.checker.GetRaftStatus()
	if err != nil {
		return nil, fmt.Errorf("getting raft status: %w", err)
	}

	result := &RaftResult{
		Leader:     status.Leader,
		TotalPeers: len(status.Servers),
	}

	for _, s := range status.Servers {
		if s.Voter {
			result.VoterCount++
		}
	}

	switch {
	case result.Leader == "":
		result.Healthy = false
		result.Message = "raft cluster has no elected leader"
		m.logger.Warn("raft cluster unhealthy: no leader")
	case result.VoterCount == 0:
		result.Healthy = false
		result.Message = "raft cluster has no voting members"
		m.logger.Warn("raft cluster unhealthy: no voters")
	default:
		result.Healthy = true
		result.Message = fmt.Sprintf("raft cluster healthy: leader=%s peers=%d voters=%d",
			result.Leader, result.TotalPeers, result.VoterCount)
		m.logger.Info("raft cluster healthy",
			zap.String("leader", result.Leader),
			zap.Int("peers", result.TotalPeers),
			zap.Int("voters", result.VoterCount),
		)
	}

	return result, nil
}
