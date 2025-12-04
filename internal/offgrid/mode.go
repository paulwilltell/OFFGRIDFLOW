// Package offgrid provides connectivity mode management for the OffGridFlow
// platform. It enables seamless switching between online (cloud) and offline
// (local) operation modes, allowing the system to gracefully degrade when
// network connectivity is unavailable.
//
// The package consists of two main components:
//
//   - ModeManager: Thread-safe state machine for tracking and transitioning
//     between online/offline modes with observer pattern support
//
//   - ConnectivityWatcher: Background service that monitors network
//     connectivity and automatically updates the ModeManager
//
// Usage:
//
//	mm := offgrid.NewModeManager(offgrid.ModeOnline)
//	mm.OnChange(func(old, new offgrid.Mode) {
//	    log.Printf("Mode changed: %s -> %s", old, new)
//	})
//
//	watcher := offgrid.NewConnectivityWatcher(mm, offgrid.DefaultWatcherConfig())
//	go watcher.Start(ctx)
package offgrid

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// Mode Type and Constants
// =============================================================================

// Mode represents the current connectivity state of the system.
// It determines how services route requests (cloud vs local).
type Mode string

const (
	// ModeOnline indicates the system has network connectivity and should
	// prefer cloud services. This is the default state when connectivity
	// is confirmed.
	ModeOnline Mode = "ONLINE"

	// ModeOffline indicates the system lacks network connectivity and
	// should use local fallback services. All cloud API calls should be
	// avoided in this mode.
	ModeOffline Mode = "OFFLINE"
)

// String returns a human-readable representation of the mode.
func (m Mode) String() string {
	return string(m)
}

// IsValid returns true if the mode is a recognized value.
func (m Mode) IsValid() bool {
	switch m {
	case ModeOnline, ModeOffline:
		return true
	default:
		return false
	}
}

// IsOnline returns true if the mode represents online connectivity.
func (m Mode) IsOnline() bool {
	return m == ModeOnline
}

// IsOffline returns true if the mode represents offline state.
func (m Mode) IsOffline() bool {
	return m == ModeOffline
}

// MarshalJSON implements json.Marshaler for Mode.
func (m Mode) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(m))
}

// UnmarshalJSON implements json.Unmarshaler for Mode.
func (m *Mode) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	mode := Mode(s)
	if !mode.IsValid() {
		return fmt.Errorf("offgrid: invalid mode: %q", s)
	}
	*m = mode
	return nil
}

// =============================================================================
// Sentinel Errors
// =============================================================================

var (
	// ErrNilModeManager is returned when a nil ModeManager is used.
	ErrNilModeManager = errors.New("offgrid: nil mode manager")

	// ErrInvalidMode is returned when an invalid mode value is provided.
	ErrInvalidMode = errors.New("offgrid: invalid mode")
)

// =============================================================================
// Mode Change Callback Type
// =============================================================================

// ModeChangeCallback is invoked when the mode transitions.
// The callback receives the previous and new modes.
//
// Important: Callbacks are invoked synchronously after the mode change.
// Long-running operations should be dispatched asynchronously to avoid
// blocking other callbacks and mode updates.
type ModeChangeCallback func(old, new Mode)

// =============================================================================
// ModeManager Implementation
// =============================================================================

// ModeManager is a thread-safe state machine for tracking connectivity mode.
// It provides an observer pattern for components to react to mode changes.
//
// Thread safety: All methods are safe for concurrent use.
//
// Example:
//
//	mm := offgrid.NewModeManager(offgrid.ModeOnline)
//
//	// Register callback before mode changes
//	mm.OnChange(func(old, new offgrid.Mode) {
//	    log.Printf("Mode transition: %s -> %s", old, new)
//	})
//
//	// Trigger mode change
//	mm.SetMode(offgrid.ModeOffline)
type ModeManager struct {
	mu        sync.RWMutex
	mode      Mode
	listeners []ModeChangeCallback

	// Metrics
	transitionCount atomic.Int64
	lastTransition  atomic.Value // time.Time
}

// NewModeManager creates a new ModeManager with the specified initial mode.
// If the mode is invalid, it defaults to ModeOnline.
func NewModeManager(initial Mode) *ModeManager {
	if !initial.IsValid() {
		initial = ModeOnline
	}

	mm := &ModeManager{
		mode:      initial,
		listeners: make([]ModeChangeCallback, 0, 4),
	}

	mm.lastTransition.Store(time.Now())
	return mm
}

// GetMode returns the current connectivity mode.
// This method is safe for concurrent use.
func (m *ModeManager) GetMode() Mode {
	if m == nil {
		return ModeOffline // Safe default for nil manager
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// SetMode transitions to the specified mode.
// If the mode is the same as the current mode, this is a no-op.
// Registered callbacks are invoked synchronously after the state change.
//
// Callbacks are invoked outside the lock to prevent deadlocks, but this means
// the mode could change again before all callbacks complete. Callbacks receive
// the old and new mode at the time of the transition.
func (m *ModeManager) SetMode(newMode Mode) {
	if m == nil {
		return
	}

	if !newMode.IsValid() {
		return // Silently ignore invalid modes
	}

	m.mu.Lock()

	// No-op if mode unchanged
	if newMode == m.mode {
		m.mu.Unlock()
		return
	}

	old := m.mode
	m.mode = newMode

	// Copy listeners slice to avoid holding lock during callbacks
	listeners := make([]ModeChangeCallback, len(m.listeners))
	copy(listeners, m.listeners)

	m.mu.Unlock()

	// Update metrics
	m.transitionCount.Add(1)
	m.lastTransition.Store(time.Now())

	// Invoke callbacks outside the lock
	for _, listener := range listeners {
		if listener != nil {
			// Wrap in recovery to prevent one bad callback from breaking others
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Log panic but continue with other callbacks
						// In production, this should use the configured logger
					}
				}()
				listener(old, newMode)
			}()
		}
	}
}

// OnChange registers a callback that will be invoked whenever the mode changes.
// Callbacks are invoked synchronously in registration order.
//
// Important notes:
//   - Callbacks should be fast; long operations should be async
//   - Callbacks may be invoked from any goroutine
//   - Multiple rapid mode changes may result in overlapping callback invocations
//   - There is no way to unregister callbacks (by design, to keep implementation simple)
func (m *ModeManager) OnChange(fn ModeChangeCallback) {
	if m == nil || fn == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, fn)
}

// IsOnline is a convenience method that returns true if currently in online mode.
func (m *ModeManager) IsOnline() bool {
	return m.GetMode().IsOnline()
}

// IsOffline is a convenience method that returns true if currently in offline mode.
func (m *ModeManager) IsOffline() bool {
	return m.GetMode().IsOffline()
}

// Stats returns current mode manager statistics.
func (m *ModeManager) Stats() ModeStats {
	if m == nil {
		return ModeStats{}
	}

	m.mu.RLock()
	mode := m.mode
	listenerCount := len(m.listeners)
	m.mu.RUnlock()

	lastTransition, _ := m.lastTransition.Load().(time.Time)

	return ModeStats{
		CurrentMode:     mode,
		TransitionCount: m.transitionCount.Load(),
		LastTransition:  lastTransition,
		ListenerCount:   listenerCount,
	}
}

// ModeStats contains statistics about the ModeManager.
type ModeStats struct {
	// CurrentMode is the current connectivity mode.
	CurrentMode Mode `json:"current_mode"`

	// TransitionCount is the total number of mode transitions since creation.
	TransitionCount int64 `json:"transition_count"`

	// LastTransition is the timestamp of the most recent mode change.
	LastTransition time.Time `json:"last_transition"`

	// ListenerCount is the number of registered change callbacks.
	ListenerCount int `json:"listener_count"`
}

// =============================================================================
// ForceMode for Testing
// =============================================================================

// ForceMode is a helper for tests that need to temporarily override the mode.
// It sets the mode and returns a cleanup function that restores the original.
//
// Usage:
//
//	cleanup := mm.ForceMode(offgrid.ModeOffline)
//	defer cleanup()
//	// ... test code that expects offline mode ...
func (m *ModeManager) ForceMode(mode Mode) (cleanup func()) {
	original := m.GetMode()
	m.SetMode(mode)
	return func() {
		m.SetMode(original)
	}
}
