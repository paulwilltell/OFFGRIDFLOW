// Package ingestion provides tests for common utilities.
package ingestion

import (
	"context"
	"testing"
	"time"
)

// TestRateLimiterBasic verifies basic token bucket functionality.
func TestRateLimiterBasic(t *testing.T) {
	limiter := NewRateLimiter(5, 10.0, 50*time.Millisecond)

	ctx := context.Background()

	// Should allow 5 tokens initially
	for i := 0; i < 5; i++ {
		_, err := limiter.Allow(ctx)
		if err != nil {
			t.Fatalf("Allow %d failed: %v", i+1, err)
		}
	}

	// 6th token should wait
	start := time.Now()
	_, err := limiter.Allow(ctx)
	if err != nil {
		t.Fatalf("Allow 6th failed: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 50*time.Millisecond {
		t.Logf("Wait time %.0fms is less than expected 100ms", elapsed.Seconds()*1000)
	}

	t.Log("✓ Rate limiter basic test passed")
}

// TestRateLimiterContextCancellation verifies that context cancellation is respected.
func TestRateLimiterContextCancellation(t *testing.T) {
	limiter := NewRateLimiter(1, 1.0, 100*time.Millisecond)

	// Consume the single token
	ctx := context.Background()
	_, _ = limiter.Allow(ctx)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Allow should respect cancellation
	_, err := limiter.Allow(ctx)
	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled, got %v", err)
	}

	t.Log("✓ Rate limiter context cancellation test passed")
}

// TestRateLimiterTryAllow verifies non-blocking token consumption.
func TestRateLimiterTryAllow(t *testing.T) {
	limiter := NewRateLimiter(2, 10.0, 50*time.Millisecond)

	// First two should succeed
	if !limiter.TryAllow() {
		t.Fatalf("TryAllow 1 should succeed")
	}
	if !limiter.TryAllow() {
		t.Fatalf("TryAllow 2 should succeed")
	}

	// Third should fail
	if limiter.TryAllow() {
		t.Fatalf("TryAllow 3 should fail")
	}

	t.Log("✓ Rate limiter TryAllow test passed")
}

// TestPaginationStateCursor verifies cursor-based pagination.
func TestPaginationStateCursor(t *testing.T) {
	pagination := NewPaginationState(100)
	pagination.SetCursorBased("cursor_1")

	if pagination.Cursor != "cursor_1" {
		t.Fatalf("Expected cursor_1, got %s", pagination.Cursor)
	}
	if !pagination.HasMore {
		t.Fatalf("HasMore should be true")
	}

	// Update to next page
	err := pagination.UpdateCursor("cursor_2", 100)
	if err != nil {
		t.Fatalf("UpdateCursor failed: %v", err)
	}

	if pagination.TotalFetched != 100 {
		t.Fatalf("Expected 100 total fetched, got %d", pagination.TotalFetched)
	}

	// Advance to end
	err = pagination.UpdateCursor("", 50)
	if err != nil {
		t.Fatalf("UpdateCursor to end failed: %v", err)
	}

	if pagination.HasMore {
		t.Fatalf("HasMore should be false after empty cursor")
	}

	t.Log("✓ Pagination cursor test passed")
}

// TestPaginationStateOffset verifies offset-based pagination.
func TestPaginationStateOffset(t *testing.T) {
	pagination := NewPaginationState(25)
	pagination.SetOffsetBased(0)

	if pagination.Offset != 0 {
		t.Fatalf("Expected offset 0, got %d", pagination.Offset)
	}

	// First page
	err := pagination.AdvancePage()
	if err != nil {
		t.Fatalf("AdvancePage 1 failed: %v", err)
	}

	nextOffset := pagination.NextOffset()
	if nextOffset != 25 { // page 1 * pagesize 25
		t.Fatalf("Expected offset 25, got %d", nextOffset)
	}

	t.Log("✓ Pagination offset test passed")
}

// TestPaginationMaxPages verifies max pages limit enforcement.
func TestPaginationMaxPages(t *testing.T) {
	pagination := NewPaginationState(100)
	pagination.MaxPages = 3
	pagination.SetCursorBased("cursor_1")

	// Advance 3 times (should hit max)
	for i := 0; i < 3; i++ {
		err := pagination.AdvancePage()
		if err != nil {
			t.Fatalf("AdvancePage %d failed: %v", i+1, err)
		}
	}

	// 4th advance should fail
	err := pagination.AdvancePage()
	if err == nil {
		t.Fatalf("Expected error on max pages exceeded")
	}

	if pagination.HasMore {
		t.Fatalf("HasMore should be false after max pages")
	}

	t.Log("✓ Pagination max pages test passed")
}

// TestErrorClassification verifies error classification.
func TestErrorClassification(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedClass ErrorClass
		shouldRetry   bool
	}{
		{
			name:          "timeout error",
			err:           newError("context deadline exceeded"),
			expectedClass: ErrorClassTransient,
			shouldRetry:   true,
		},
		{
			name:          "rate limit error",
			err:           newError("rate limit exceeded (429)"),
			expectedClass: ErrorClassTransient,
			shouldRetry:   true,
		},
		{
			name:          "auth error",
			err:           newError("unauthorized (401)"),
			expectedClass: ErrorClassAuth,
			shouldRetry:   false,
		},
		{
			name:          "not found error",
			err:           newError("bucket not found (404)"),
			expectedClass: ErrorClassNotFound,
			shouldRetry:   false,
		},
		{
			name:          "bad request",
			err:           newError("bad request: invalid JSON"),
			expectedClass: ErrorClassBadRequest,
			shouldRetry:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := ClassifyError(tt.err)
			if ce.Class != tt.expectedClass {
				t.Errorf("Expected %s, got %s", tt.expectedClass, ce.Class)
			}
			if ce.IsRetryable() != tt.shouldRetry {
				t.Errorf("Expected retryable=%v, got %v", tt.shouldRetry, ce.IsRetryable())
			}
		})
	}

	t.Log("✓ Error classification test passed")
}

// TestHTTPErrorClassification verifies HTTP status code classification.
func TestHTTPErrorClassification(t *testing.T) {
	tests := []struct {
		status        int
		expectedClass ErrorClass
	}{
		{400, ErrorClassBadRequest},
		{401, ErrorClassAuth},
		{403, ErrorClassAuth},
		{404, ErrorClassNotFound},
		{429, ErrorClassTransient},
		{500, ErrorClassFatal},
		{503, ErrorClassTransient},
		{504, ErrorClassTransient},
	}

	for _, tt := range tests {
		ce := ClassifyHTTPError(tt.status, "")
		if ce.Class != tt.expectedClass {
			t.Errorf("Status %d: expected %s, got %s", tt.status, tt.expectedClass, ce.Class)
		}
	}

	t.Log("✓ HTTP error classification test passed")
}

// TestShouldRetry verifies retry decision logic.
func TestShouldRetry(t *testing.T) {
	retryableErr := newError("temporary failure in service")
	nonRetryableErr := newError("unauthorized")

	if !ShouldRetry(retryableErr) {
		t.Fatalf("Should retry for transient error")
	}

	if ShouldRetry(nonRetryableErr) {
		t.Fatalf("Should not retry for auth error")
	}

	t.Log("✓ Should retry test passed")
}

// newError creates a simple test error.
func newError(msg string) error {
	return errorType{msg: msg}
}

type errorType struct {
	msg string
}

func (e errorType) Error() string {
	return e.msg
}
