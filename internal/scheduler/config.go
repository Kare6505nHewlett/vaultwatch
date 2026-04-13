package scheduler

import (
	"fmt"
	"time"

	"github.com/yourusername/vaultwatch/internal/config"
)

// JobsFromConfig converts the application config into a list of scheduler Jobs.
func JobsFromConfig(cfg *config.Config) ([]Job, error) {
	if len(cfg.Secrets) == 0 {
		return nil, fmt.Errorf("scheduler: no secrets configured")
	}

	interval, err := time.ParseDuration(cfg.CheckInterval)
	if err != nil {
		return nil, fmt.Errorf("scheduler: invalid check_interval %q: %w", cfg.CheckInterval, err)
	}

	if interval <= 0 {
		return nil, fmt.Errorf("scheduler: check_interval must be positive, got %s", interval)
	}

	jobs := make([]Job, 0, len(cfg.Secrets))
	for _, s := range cfg.Secrets {
		if s.Path == "" {
			return nil, fmt.Errorf("scheduler: secret entry has empty path")
		}
		jobs = append(jobs, Job{
			SecretPath: s.Path,
			Interval:   interval,
		})
	}

	return jobs, nil
}
