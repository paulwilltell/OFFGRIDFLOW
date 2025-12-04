package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Orchestrator runs ingestion with retry logic and optional observability.
type Orchestrator struct {
	Service        *Service
	Attempts       int
	InitialBackoff time.Duration
	Logger         *slog.Logger
}

// Run executes the ingestion service with retries and returns the processed activities.
func (o *Orchestrator) Run(ctx context.Context) ([]Activity, error) {
	if o == nil || o.Service == nil {
		return nil, fmt.Errorf("orchestrator: ingestion service not configured")
	}
	attempts := o.Attempts
	if attempts <= 0 {
		attempts = 3
	}
	backoff := o.InitialBackoff
	if backoff <= 0 {
		backoff = 2 * time.Second
	}

	var activities []Activity
	err := WithRetry(ctx, attempts, backoff, func() error {
		var err error
		activities, err = o.Service.Run(ctx)
		return err
	})
	if err != nil {
		if o.Logger != nil {
			o.Logger.Warn("orchestrator failed", "error", err)
		}
		return nil, fmt.Errorf("orchestrator: %w", err)
	}

	if o.Logger != nil {
		o.Logger.Info("orchestrator completed", "count", len(activities))
	}

	return activities, nil
}
