package offgrid

import (
	"testing"
)

func TestMustNewConnectivityWatcher_NilPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewConnectivityWatcher should panic with nil ModeManager")
		} else {
			// Verify panic message
			msg, ok := r.(string)
			if !ok || msg != "offgrid: MustNewConnectivityWatcher requires non-nil ModeManager" {
				t.Errorf("Unexpected panic message: %v", r)
			}
		}
	}()

	// This should panic
	MustNewConnectivityWatcher(nil, WatcherConfig{})
}
