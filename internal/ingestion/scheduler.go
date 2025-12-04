package ingestion

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Scheduler triggers orchestrated ingestion runs on a fixed interval.
type Scheduler struct {
	Orchestrator *Orchestrator
	Interval     time.Duration
	Logger       *slog.Logger

	mu      sync.RWMutex
	lastRun time.Time
}

// SchedulerStatus exposes the current schedule metadata.
type SchedulerStatus struct {
	LastRun  time.Time `json:"last_run_at,omitempty"`
	NextRun  time.Time `json:"next_run_at,omitempty"`
	Interval string    `json:"interval,omitempty"`
}

// Start begins the scheduler loop. It runs once immediately and thereafter on the interval.
func (s *Scheduler) Start(ctx context.Context) {
	if s == nil || s.Orchestrator == nil || s.Interval <= 0 {
		return
	}

	if err := s.execute(ctx); err != nil {
		s.logWarning("initial scheduled run failed", err)
	}

	ticker := time.NewTicker(s.Interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.execute(ctx); err != nil {
					s.logWarning("scheduled ingestion failed", err)
				}
			}
		}
	}()
}

// Status returns last/next run timestamps along with the configured interval.
func (s *Scheduler) Status() SchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status := SchedulerStatus{
		LastRun:  s.lastRun,
		Interval: s.Interval.String(),
	}
	if !s.lastRun.IsZero() {
		status.NextRun = s.lastRun.Add(s.Interval)
	}
	return status
}

func (s *Scheduler) execute(ctx context.Context) error {
	activities, err := s.Orchestrator.Run(ctx)
	now := time.Now().UTC()
	s.mu.Lock()
	s.lastRun = now
	s.mu.Unlock()

	if err != nil {
		return err
	}

	if s.Logger != nil {
		s.Logger.Info("scheduled ingestion completed", "count", len(activities))
	}
	return nil
}

func (s *Scheduler) logWarning(msg string, err error) {
	if s.Logger != nil {
		s.Logger.Warn(msg, "error", err)
	}
}
