package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/monitor"
)

// Job represents a scheduled secret check.
type Job struct {
	SecretPath string
	Interval   time.Duration
}

// Scheduler periodically checks secrets and sends alerts.
type Scheduler struct {
	checker  *monitor.Checker
	notifier *alert.LogNotifier
	jobs     []Job
	logger   *log.Logger
}

// New creates a new Scheduler.
func New(checker *monitor.Checker, notifier *alert.LogNotifier, jobs []Job, logger *log.Logger) *Scheduler {
	return &Scheduler{
		checker:  checker,
		notifier: notifier,
		jobs:     jobs,
		logger:   logger,
	}
}

// Run starts all scheduled jobs and blocks until the context is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	for _, job := range s.jobs {
		go s.runJob(ctx, job)
	}
	<-ctx.Done()
	s.logger.Println("scheduler: shutting down")
}

// runJob executes a single job on its interval until the context is cancelled.
func (s *Scheduler) runJob(ctx context.Context, job Job) {
	s.logger.Printf("scheduler: starting job for secret %s every %s", job.SecretPath, job.Interval)
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	// Run immediately on start.
	s.check(job.SecretPath)

	for {
		select {
		case <-ticker.C:
			s.check(job.SecretPath)
		case <-ctx.Done():
			s.logger.Printf("scheduler: stopping job for secret %s", job.SecretPath)
			return
		}
	}
}

// check performs a single secret expiry check and sends an alert if needed.
func (s *Scheduler) check(secretPath string) {
	result, err := s.checker.CheckSecret(secretPath)
	if err != nil {
		s.logger.Printf("scheduler: error checking secret %s: %v", secretPath, err)
		return
	}
	if err := s.notifier.Send(result); err != nil {
		s.logger.Printf("scheduler: error sending alert for %s: %v", secretPath, err)
	}
}
