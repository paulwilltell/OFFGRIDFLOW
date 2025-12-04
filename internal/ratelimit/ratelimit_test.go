package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	cfg := DefaultConfig()
	rl := NewRateLimiter(cfg)
	defer rl.Close()

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	cfg := Config{
		RequestsPerSecond: 10,
		BurstSize:         5,
		CleanupInterval:   time.Minute,
		BucketTTL:         time.Minute,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Close()

	ctx := context.Background()
	key := "test-key"

	// First 5 requests should be allowed (burst)
	for i := 0; i < 5; i++ {
		if !rl.Allow(ctx, key) {
			t.Errorf("Request %d should be allowed (within burst)", i+1)
		}
	}
}

func TestRateLimiter_Remaining(t *testing.T) {
	cfg := Config{
		RequestsPerSecond: 10,
		BurstSize:         10,
		CleanupInterval:   time.Minute,
		BucketTTL:         time.Minute,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Close()

	ctx := context.Background()
	key := "remaining-test"

	// New key should have full burst remaining
	remaining := rl.Remaining(ctx, key)
	if remaining != 10 {
		t.Errorf("Expected 10 remaining, got %d", remaining)
	}

	// Use some tokens
	rl.Allow(ctx, key)
	rl.Allow(ctx, key)

	remaining = rl.Remaining(ctx, key)
	if remaining > 10 {
		t.Errorf("Remaining should be <= 10, got %d", remaining)
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	cfg := Config{
		RequestsPerSecond: 10,
		BurstSize:         3,
		CleanupInterval:   time.Minute,
		BucketTTL:         time.Minute,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Close()

	ctx := context.Background()

	// Each key should have its own bucket
	for i := 0; i < 3; i++ {
		if !rl.Allow(ctx, "key1") {
			t.Errorf("key1 request %d should be allowed", i+1)
		}
		if !rl.Allow(ctx, "key2") {
			t.Errorf("key2 request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	cfg := Config{
		RequestsPerSecond: 100,
		BurstSize:         100,
		CleanupInterval:   time.Minute,
		BucketTTL:         time.Minute,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Close()

	ctx := context.Background()
	key := "concurrent-test"

	var wg sync.WaitGroup
	allowed := make(chan bool, 200)

	// Concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- rl.Allow(ctx, key)
		}()
	}

	wg.Wait()
	close(allowed)

	count := 0
	for a := range allowed {
		if a {
			count++
		}
	}

	// At least some should be allowed
	if count == 0 {
		t.Error("Expected at least some requests to be allowed")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.RequestsPerSecond <= 0 {
		t.Error("RequestsPerSecond should be positive")
	}
	if cfg.BurstSize <= 0 {
		t.Error("BurstSize should be positive")
	}
	if cfg.CleanupInterval <= 0 {
		t.Error("CleanupInterval should be positive")
	}
	if cfg.BucketTTL <= 0 {
		t.Error("BucketTTL should be positive")
	}
}

func TestRateLimiter_Close(t *testing.T) {
	cfg := DefaultConfig()
	rl := NewRateLimiter(cfg)

	// Should not panic and return nil
	err := rl.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}
