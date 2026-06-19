package gnolang

import (
	"testing"
)

func TestNopCoverageTracker(t *testing.T) {
	tracker := &NopCoverageTracker{}

	// Test that all methods are no-ops and don't panic
	tracker.TrackExecution("pkg", "file", 1)
	tracker.TrackStatement(nil)
	tracker.TrackExpression(nil)

	if tracker.IsEnabled() {
		t.Error("NopCoverageTracker should always return false for IsEnabled()")
	}

	tracker.SetEnabled(true) // Should be a no-op
	if tracker.IsEnabled() {
		t.Error("NopCoverageTracker should still return false after SetEnabled(true)")
	}
}

func TestDefaultCoverageTracker(t *testing.T) {
	tracker := DefaultCoverageTracker()

	if tracker == nil {
		t.Fatal("DefaultCoverageTracker() should not return nil")
	}

	// Should return a NopCoverageTracker
	if _, ok := tracker.(*NopCoverageTracker); !ok {
		t.Error("DefaultCoverageTracker() should return a *NopCoverageTracker")
	}
}
