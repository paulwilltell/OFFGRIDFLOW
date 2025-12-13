package ratelimit

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// getEnvInt reads an integer from an environment variable, or returns the default.
func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			return i
		}
	}
	return defaultVal
}

// RateLimiter provides token bucket rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*bucket
	config   Config
	ticker   *time.Ticker
	stopChan chan struct{}
}

// Config holds rate limiter configuration
type Config struct {
	RequestsPerSecond int
	BurstSize         int
	CleanupInterval   time.Duration
	BucketTTL         time.Duration
}

// bucket represents a token bucket for a single key
type bucket struct {
	tokens    float64
	lastCheck time.Time
	mu        sync.Mutex
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() Config {
	return Config{
		RequestsPerSecond: 10,
		BurstSize:         20,
		CleanupInterval:   1 * time.Minute,
		BucketTTL:         5 * time.Minute,
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config Config) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		config:   config,
		ticker:   time.NewTicker(config.CleanupInterval),
		stopChan: make(chan struct{}),
	}

	go rl.cleanup()
	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(ctx context.Context, key string) bool {
	rl.mu.Lock()
	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:    float64(rl.config.BurstSize),
			lastCheck: time.Now(),
		}
		rl.buckets[key] = b
	}
	rl.mu.Unlock()

	return b.allow(rl.config.RequestsPerSecond, rl.config.BurstSize)
}

// Remaining returns requests remaining for the key
func (rl *RateLimiter) Remaining(ctx context.Context, key string) int64 {
	rl.mu.RLock()
	b, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		return int64(rl.config.BurstSize)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * float64(rl.config.RequestsPerSecond)
	if b.tokens > float64(rl.config.BurstSize) {
		b.tokens = float64(rl.config.BurstSize)
	}
	b.lastCheck = now

	return int64(b.tokens)
}

// Close stops the rate limiter
func (rl *RateLimiter) Close() error {
	close(rl.stopChan)
	rl.ticker.Stop()
	return nil
}

func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.ticker.C:
			rl.removeExpiredBuckets()
		case <-rl.stopChan:
			return
		}
	}
}

func (rl *RateLimiter) removeExpiredBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.config.BucketTTL)
	for key, b := range rl.buckets {
		b.mu.Lock()
		if b.lastCheck.Before(cutoff) {
			delete(rl.buckets, key)
		}
		b.mu.Unlock()
	}
}

func (b *bucket) allow(rps, burst int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * float64(rps)
	if b.tokens > float64(burst) {
		b.tokens = float64(burst)
	}
	b.lastCheck = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}

	return false
}

// MultiTierLimiter applies different limits based on tiers
type MultiTierLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
}

// DefaultTiers returns default tier configurations.
// Override via env vars: RATELIMIT_FREE_RPS, RATELIMIT_FREE_BURST, etc.
func DefaultTiers() map[string]Config {
	return map[string]Config{
		"free": {
			RequestsPerSecond: getEnvInt("RATELIMIT_FREE_RPS", 5),
			BurstSize:         getEnvInt("RATELIMIT_FREE_BURST", 5),
			CleanupInterval:   1 * time.Minute,
			BucketTTL:         5 * time.Minute,
		},
		"pro": {
			RequestsPerSecond: getEnvInt("RATELIMIT_PRO_RPS", 50),
			BurstSize:         getEnvInt("RATELIMIT_PRO_BURST", 100),
			CleanupInterval:   1 * time.Minute,
			BucketTTL:         5 * time.Minute,
		},
		"enterprise": {
			RequestsPerSecond: getEnvInt("RATELIMIT_ENTERPRISE_RPS", 500),
			BurstSize:         getEnvInt("RATELIMIT_ENTERPRISE_BURST", 1000),
			CleanupInterval:   1 * time.Minute,
			BucketTTL:         5 * time.Minute,
		},
	}
}

// NewMultiTierLimiter creates a multi-tier rate limiter
func NewMultiTierLimiter(tiers map[string]Config) *MultiTierLimiter {
	limiters := make(map[string]*RateLimiter)
	for tier, config := range tiers {
		limiters[tier] = NewRateLimiter(config)
	}

	return &MultiTierLimiter{limiters: limiters}
}

// Allow checks if a request should be allowed for the tier
func (m *MultiTierLimiter) Allow(ctx context.Context, tier, key string) bool {
	m.mu.RLock()
	limiter, exists := m.limiters[tier]
	m.mu.RUnlock()

	if !exists {
		m.mu.RLock()
		limiter = m.limiters["free"]
		m.mu.RUnlock()
	}

	if limiter == nil {
		return true
	}

	return limiter.Allow(ctx, key)
}

// Remaining returns the remaining requests allowed for the tier and key
func (m *MultiTierLimiter) Remaining(ctx context.Context, tier, key string) int64 {
	m.mu.RLock()
	limiter, exists := m.limiters[tier]
	m.mu.RUnlock()

	if !exists {
		m.mu.RLock()
		limiter = m.limiters["free"]
		m.mu.RUnlock()
	}

	if limiter == nil {
		return -1 // unlimited
	}

	return limiter.Remaining(ctx, key)
}

// Close closes all tier limiters
func (m *MultiTierLimiter) Close() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, limiter := range m.limiters {
		if err := limiter.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions for key generation
func DefaultKeyFunc(tenantID string) string {
	return fmt.Sprintf("tenant:%s", tenantID)
}

func UserKeyFunc(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

func APIKeyFunc(keyID string) string {
	return fmt.Sprintf("apikey:%s", keyID)
}

func IPKeyFunc(ip string) string {
	return fmt.Sprintf("ip:%s", ip)
}
