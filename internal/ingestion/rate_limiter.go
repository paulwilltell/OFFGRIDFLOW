// Package ingestion provides rate limiting for cloud API calls.
// Implements a token bucket algorithm with jitter to prevent thundering herd.
package ingestion

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter enforces request rate limits using token bucket algorithm.
// Thread-safe for concurrent use across multiple goroutines.
type RateLimiter struct {
	mu          sync.Mutex
	tokens      float64      // Current tokens available
	capacity    float64      // Maximum tokens
	refillRate  float64      // Tokens per second
	lastRefill  time.Time    // Last time tokens were refilled
	minWaitTime time.Duration // Minimum wait time to prevent busy-loop
}

// NewRateLimiter creates a new rate limiter with specified capacity and refill rate.
//
// Args:
//   - capacity: maximum tokens (e.g., 100)
//   - refillRate: tokens per second (e.g., 10.0 for 10 requests/sec)
//   - minWaitTime: minimum wait time between attempts (prevent CPU spin)
//
// Example: Allow 100 requests with 10 requests/second refill:
//   limiter := NewRateLimiter(100, 10.0, 100*time.Millisecond)
func NewRateLimiter(capacity, refillRate float64, minWaitTime time.Duration) *RateLimiter {
	if capacity <= 0 {
		capacity = 100
	}
	if refillRate <= 0 {
		refillRate = 10.0
	}
	if minWaitTime <= 0 {
		minWaitTime = 100 * time.Millisecond
	}

	return &RateLimiter{
		tokens:      capacity,
		capacity:    capacity,
		refillRate:  refillRate,
		lastRefill:  time.Now(),
		minWaitTime: minWaitTime,
	}
}

// Allow waits until a token is available, then consumes it.
// Returns the time waited and any context errors.
//
// Context cancellation is respected: if ctx is cancelled before a token
// becomes available, returns immediately with ctx.Err().
func (rl *RateLimiter) Allow(ctx context.Context) (time.Duration, error) {
	start := time.Now()

	for {
		rl.mu.Lock()
		rl.refill()

		if rl.tokens >= 1 {
			rl.tokens--
			rl.mu.Unlock()
			return time.Since(start), nil
		}

		// Calculate time until next token is available
		tokensNeeded := 1 - rl.tokens
		waitDuration := time.Duration(float64(time.Second) * tokensNeeded / rl.refillRate)
		if waitDuration < rl.minWaitTime {
			waitDuration = rl.minWaitTime
		}
		rl.mu.Unlock()

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return time.Since(start), ctx.Err()
		case <-time.After(waitDuration):
			// Try again
		}
	}
}

// AllowN waits for n tokens and consumes them.
// Returns error if n > capacity.
func (rl *RateLimiter) AllowN(ctx context.Context, n int) (time.Duration, error) {
	if float64(n) > rl.capacity {
		return 0, fmt.Errorf("rate_limit: requested %d tokens exceeds capacity of %.0f", n, rl.capacity)
	}

	start := time.Now()
	for i := 0; i < n; i++ {
		_, err := rl.Allow(ctx)
		if err != nil {
			return time.Since(start), err
		}
	}
	return time.Since(start), nil
}

// TryAllow attempts to consume a token without waiting.
// Returns true if successful, false if no tokens available.
func (rl *RateLimiter) TryAllow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.refill()

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

// refill adds tokens based on time elapsed since last refill.
// Must be called while holding mutex.
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	tokensToAdd := elapsed * rl.refillRate

	rl.tokens = min(rl.tokens+tokensToAdd, rl.capacity)
	rl.lastRefill = now
}

// Reset restores the limiter to full capacity.
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.tokens = rl.capacity
	rl.lastRefill = time.Now()
}

// Available returns the current number of available tokens (non-blocking).
func (rl *RateLimiter) Available() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.refill()
	return rl.tokens
}

// min returns the smaller of two values.
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
