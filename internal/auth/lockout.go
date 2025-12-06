package auth

import (
	"strings"
	"sync"
	"time"
)

// LoginAttempt tracks recent failures for an identity.
type LoginAttempt struct {
	Email        string
	Attempts     int
	FirstAttempt time.Time
	LockedUntil  *time.Time
}

// LockoutManager tracks failed authentication attempts and enforces temporary locks.
type LockoutManager struct {
	mu             sync.RWMutex
	attempts       map[string]*LoginAttempt
	maxAttempts    int
	lockoutPeriod  time.Duration
	windowDuration time.Duration
}

// NewLockoutManager creates a lockout manager.
// maxAttempts controls how many failed logins are allowed before a lockout.
// lockoutDuration sets how long the account remains locked.
// windowDuration resets the counter if failed attempts are spread outside this window.
// cleanupInterval determines how often expired records are purged.
func NewLockoutManager(maxAttempts int, lockoutDuration, windowDuration, cleanupInterval time.Duration) *LockoutManager {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if lockoutDuration <= 0 {
		lockoutDuration = 15 * time.Minute
	}
	if windowDuration <= 0 {
		windowDuration = 5 * time.Minute
	}
	if cleanupInterval <= 0 {
		cleanupInterval = 5 * time.Minute
	}

	m := &LockoutManager{
		attempts:       make(map[string]*LoginAttempt),
		maxAttempts:    maxAttempts,
		lockoutPeriod:  lockoutDuration,
		windowDuration: windowDuration,
	}

	go m.cleanup(cleanupInterval)

	return m
}

// IsLocked reports whether the given email is currently locked.
func (m *LockoutManager) IsLocked(email string) bool {
	key := m.normalize(email)
	m.mu.RLock()
	defer m.mu.RUnlock()

	attempt, ok := m.attempts[key]
	if !ok || attempt.LockedUntil == nil {
		return false
	}

	return time.Now().Before(*attempt.LockedUntil)
}

// RecordFailure increments the failure counter and locks the account if needed.
// Returns whether the account is now locked and how many attempts remain before lockout.
func (m *LockoutManager) RecordFailure(email string) (bool, int) {
	key := m.normalize(email)
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	attempt, ok := m.attempts[key]
	if !ok {
		m.attempts[key] = &LoginAttempt{
			Email:        key,
			Attempts:     1,
			FirstAttempt: now,
		}
		return false, m.maxAttempts - 1
	}

	if now.Sub(attempt.FirstAttempt) > m.windowDuration {
		attempt.Attempts = 1
		attempt.FirstAttempt = now
		attempt.LockedUntil = nil
		return false, m.maxAttempts - 1
	}

	attempt.Attempts++

	if attempt.Attempts >= m.maxAttempts {
		lockUntil := now.Add(m.lockoutPeriod)
		attempt.LockedUntil = &lockUntil
		return true, 0
	}

	remaining := m.maxAttempts - attempt.Attempts
	if remaining < 0 {
		remaining = 0
	}
	return false, remaining
}

// RecordSuccess clears tracking for the provided email.
func (m *LockoutManager) RecordSuccess(email string) {
	key := m.normalize(email)
	m.mu.Lock()
	delete(m.attempts, key)
	m.mu.Unlock()
}

// GetLockoutInfo returns the number of recorded attempts and when the lockout expires.
func (m *LockoutManager) GetLockoutInfo(email string) (int, *time.Time) {
	key := m.normalize(email)
	m.mu.RLock()
	defer m.mu.RUnlock()

	attempt, ok := m.attempts[key]
	if !ok {
		return 0, nil
	}
	return attempt.Attempts, attempt.LockedUntil
}

func (m *LockoutManager) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for key, attempt := range m.attempts {
			if attempt.LockedUntil != nil && now.After(*attempt.LockedUntil) {
				delete(m.attempts, key)
				continue
			}
			if attempt.LockedUntil == nil && now.Sub(attempt.FirstAttempt) > m.windowDuration {
				delete(m.attempts, key)
			}
		}
		m.mu.Unlock()
	}
}

func (m *LockoutManager) normalize(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
