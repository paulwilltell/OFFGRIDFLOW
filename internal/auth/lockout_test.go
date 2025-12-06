package auth

import (
	"testing"
	"time"
)

func TestLockoutManager_RecordFailure(t *testing.T) {
	lm := NewLockoutManager(3, time.Minute, time.Minute, time.Hour)
	email := "user@example.com"

	// First failure should not lock and should report remaining attempts
	locked, remaining := lm.RecordFailure(email)
	if locked {
		t.Fatalf("expected not locked after first failure")
	}
	if remaining != 2 {
		t.Fatalf("expected 2 remaining attempts, got %d", remaining)
	}

	// Second failure
	locked, remaining = lm.RecordFailure(email)
	if locked {
		t.Fatalf("expected not locked after second failure")
	}
	if remaining != 1 {
		t.Fatalf("expected 1 remaining attempt, got %d", remaining)
	}

	// Third failure should lock
	locked, remaining = lm.RecordFailure(email)
	if !locked {
		t.Fatalf("expected locked after max attempts")
	}
	if remaining != 0 {
		t.Fatalf("expected 0 remaining attempts, got %d", remaining)
	}

	if !lm.IsLocked(email) {
		t.Fatalf("expected email to be locked")
	}
}

func TestLockoutManager_RecordSuccessClearsAttempts(t *testing.T) {
	lm := NewLockoutManager(3, time.Minute, time.Minute, time.Hour)
	email := "user@example.com"

	lm.RecordFailure(email)
	lm.RecordFailure(email)
	lm.RecordSuccess(email)

	if lm.IsLocked(email) {
		t.Fatalf("expected not locked after success")
	}

	attempts, _ := lm.GetLockoutInfo(email)
	if attempts != 0 {
		t.Fatalf("expected attempts to reset, got %d", attempts)
	}
}

func TestLockoutManager_WindowResetsAfterDuration(t *testing.T) {
	lm := NewLockoutManager(3, time.Minute, 1*time.Second, time.Hour)
	email := "user@example.com"

	lm.RecordFailure(email)

	key := lm.normalize(email)
	lm.mu.Lock()
	attempt := lm.attempts[key]
	attempt.FirstAttempt = time.Now().Add(-2 * time.Second)
	lm.mu.Unlock()

	locked, remaining := lm.RecordFailure(email)
	if locked {
		t.Fatalf("expected not locked after window reset")
	}
	if remaining != 2 {
		t.Fatalf("expected counter reset, got %d", remaining)
	}
}
