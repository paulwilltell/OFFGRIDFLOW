package worker

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

// Job represents a unit of background work.
type Job interface {
	// Name returns a short identifier for logging/metrics.
	Name() string
	// Run executes the job logic. Implementations should respect ctx cancellation.
	Run(ctx context.Context) error
}

// JobSpec describes how a job should be scheduled and retried.
type JobSpec struct {
	Job     Job
	Tags    []string
	Every   time.Duration
	Timeout time.Duration

	// RetryLimit is the maximum number of retry attempts per run (0 = no retries).
	RetryLimit int
	// BackoffInitial is the starting backoff when retrying.
	BackoffInitial time.Duration
	// BackoffMax caps the backoff delay.
	BackoffMax time.Duration
	// Jitter randomizes the schedule to avoid thundering herd.
	Jitter time.Duration
}

// Runner executes scheduled jobs with retry/backoff.
type Runner struct {
	logger  *slog.Logger
	specs   []JobSpec
	wg      sync.WaitGroup
	seeded  bool
	metrics *MetricsRecorder
	alerts  *AlertQueue
}

// NewRunner creates a Runner.
func NewRunner(logger *slog.Logger, specs []JobSpec, metrics *MetricsRecorder, alerts *AlertQueue) *Runner {
	return &Runner{
		logger:  logger,
		specs:   specs,
		seeded:  false,
		metrics: metrics,
		alerts:  alerts,
	}
}

// Start begins scheduling all jobs until ctx is cancelled.
func (r *Runner) Start(ctx context.Context) {
	r.seedRandOnce()
	for _, spec := range r.specs {
		if spec.Job == nil || spec.Every <= 0 {
			continue
		}
		spec := spec // capture loop variable
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			r.runJobLoop(ctx, spec)
		}()
	}
}

// Wait waits for all job loops to exit.
func (r *Runner) Wait() {
	r.wg.Wait()
}

func (r *Runner) runJobLoop(ctx context.Context, spec JobSpec) {
	logger := r.logger.With("job", spec.Job.Name())

	// Run immediately on startup to avoid waiting a full interval.
	r.runOnce(ctx, logger, spec)

	interval := spec.Every
	ticker := time.NewTicker(interval + r.randomJitter(spec.Jitter))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker loop stopping", "reason", "context_cancelled")
			return
		case <-ticker.C:
			r.runOnce(ctx, logger, spec)
			ticker.Reset(interval + r.randomJitter(spec.Jitter))
		}
	}
}

func (r *Runner) runOnce(ctx context.Context, logger *slog.Logger, spec JobSpec) {
	start := time.Now()
	r.metrics.RecordStart(ctx, spec.Job.Name())

	runCtx := ctx
	cancel := func() {}
	if spec.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, spec.Timeout)
	}
	defer cancel()

	var err error
	backoff := spec.BackoffInitial
	if backoff <= 0 {
		backoff = 500 * time.Millisecond
	}
	if spec.BackoffMax <= 0 {
		spec.BackoffMax = 30 * time.Second
	}

	attempts := 0
	for {
		attempts++
		err = spec.Job.Run(runCtx)
		if err == nil {
			logger.Info("job completed", "attempts", attempts, "duration", time.Since(start))
			r.metrics.RecordSuccess(ctx, spec.Job.Name(), time.Since(start))
			return
		}

		if runCtx.Err() != nil {
			logger.Warn("job cancelled", "attempts", attempts, "duration", time.Since(start), "error", err)
			r.metrics.RecordFailure(ctx, spec.Job.Name(), time.Since(start))
			return
		}

		if attempts > spec.RetryLimit && spec.RetryLimit >= 0 {
			logger.Error("job failed after retries", "attempts", attempts, "duration", time.Since(start), "error", err)
			r.metrics.RecordFailure(ctx, spec.Job.Name(), time.Since(start))
			r.publishAlert(spec.Job.Name(), err)
			return
		}

		sleep := backoff + r.randomJitter(spec.Jitter/2)
		if sleep > spec.BackoffMax {
			sleep = spec.BackoffMax
		}
		logger.Warn("job failed, retrying", "attempt", attempts, "sleep", sleep, "error", err)
		r.metrics.RecordRetry(ctx, spec.Job.Name())
		select {
		case <-runCtx.Done():
			logger.Warn("job cancelled before retry", "attempts", attempts, "error", err)
			r.metrics.RecordFailure(ctx, spec.Job.Name(), time.Since(start))
			return
		case <-time.After(sleep):
		}
		backoff *= 2
		if backoff > spec.BackoffMax {
			backoff = spec.BackoffMax
		}
	}
}

func (r *Runner) publishAlert(job string, err error) {
	if r.alerts == nil || err == nil {
		return
	}
	r.alerts.Publish(Alert{
		Time:     time.Now().UTC(),
		Job:      job,
		Severity: "error",
		Message:  "job failed after retries",
		Error:    err.Error(),
	})
}

func (r *Runner) randomJitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(max)))
}

func (r *Runner) seedRandOnce() {
	// Note: rand.Seed is deprecated in Go 1.20+, global rand is automatically seeded
	r.seeded = true
}
