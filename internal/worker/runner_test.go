package worker

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

type testJob struct {
	name     string
	count    atomic.Int32
	failOnce atomic.Bool
}

func (j *testJob) Name() string { return j.name }

func (j *testJob) Run(ctx context.Context) error {
	j.count.Add(1)
	if j.failOnce.CompareAndSwap(false, true) {
		return context.DeadlineExceeded
	}
	<-time.After(5 * time.Millisecond) // small work
	return nil
}

func TestRunnerExecutesAndRetries(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	job := &testJob{name: "test_job"}

	r := NewRunner(logger, []JobSpec{
		{
			Job:            job,
			Every:          20 * time.Millisecond,
			Timeout:        200 * time.Millisecond,
			RetryLimit:     1,
			BackoffInitial: 5 * time.Millisecond,
			BackoffMax:     10 * time.Millisecond,
			Jitter:         2 * time.Millisecond,
		},
	}, nil, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	r.Start(ctx)
	r.Wait()

	if attempts := job.count.Load(); attempts < 2 {
		t.Fatalf("expected job to run with retry, got %d attempts", attempts)
	}
}
