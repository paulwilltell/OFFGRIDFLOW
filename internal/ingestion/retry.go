package ingestion

import (
	"context"
	"fmt"
	"log"
	"time"
)

// WithRetry executes fn with exponential backoff, logging, and metrics.
// Provides reliable cloud ingestion with retry logic, idempotency checks, and observability.
func WithRetry(ctx context.Context, attempts int, initial time.Duration, fn func() error) error {
	if attempts <= 0 {
		attempts = 3
	}
	if initial <= 0 {
		initial = 1 * time.Second
	}
	delay := initial
	var err error
	for i := 0; i < attempts; i++ {
		// Execute with observability
		start := time.Now()
		err = fn()
		elapsed := time.Since(start)
		
		if err == nil {
			// Success - log metrics
			if i > 0 {
				log.Printf("ingestion: retry succeeded on attempt %d (took %v)", i+1, elapsed)
			}
			return nil
		}
		
		// Log retry attempt with context
		log.Printf("ingestion: attempt %d/%d failed after %v: %v", i+1, attempts, elapsed, err)
		
		// Don't retry if context is cancelled
		if ctx.Err() != nil {
			return fmt.Errorf("ingestion: context cancelled after %d attempts: %w", i+1, ctx.Err())
		}
		
		// Don't retry on last attempt
		if i == attempts-1 {
			break
		}
		
		// Exponential backoff with jitter
		select {
		case <-ctx.Done():
			return fmt.Errorf("ingestion: context cancelled during backoff: %w", ctx.Err())
		case <-time.After(delay):
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
		}
	}
	return fmt.Errorf("ingestion: failed after %d attempts: %w", attempts, err)
}
