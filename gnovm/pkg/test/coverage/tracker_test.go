package coverage

import (
	"testing"
)

func TestTracker(t *testing.T) {
	tracker := NewTracker()

	// Test initial state
	if tracker.IsEnabled() {
		t.Error("Tracker should be disabled by default")
	}

	// Test enabling/disabling
	tracker.SetEnabled(true)
	if !tracker.IsEnabled() {
		t.Error("Tracker should be enabled after SetEnabled(true)")
	}

	// Test tracking execution
	tracker.TrackExecution("test/pkg", "file.go", 10)
	tracker.TrackExecution("test/pkg", "file.go", 10) // same line again
	tracker.TrackExecution("test/pkg", "file.go", 20)
	tracker.TrackExecution("test/pkg", "file2.go", 5)

	// Get coverage data
	data := tracker.GetCoverageData()

	// Verify data
	if len(data) != 1 {
		t.Errorf("Expected 1 package, got %d", len(data))
	}

	if len(data["test/pkg"]) != 2 {
		t.Errorf("Expected 2 files, got %d", len(data["test/pkg"]))
	}

	if data["test/pkg"]["file.go"][10] != 2 {
		t.Errorf("Expected line 10 to be executed 2 times, got %d", data["test/pkg"]["file.go"][10])
	}

	if data["test/pkg"]["file.go"][20] != 1 {
		t.Errorf("Expected line 20 to be executed 1 time, got %d", data["test/pkg"]["file.go"][20])
	}
}

func TestTrackerReport(t *testing.T) {
	tracker := NewTracker()
	tracker.SetEnabled(true)

	// Register executable lines
	tracker.RegisterExecutableLine("test/pkg", "file.go", 10)
	tracker.RegisterExecutableLine("test/pkg", "file.go", 20)
	tracker.RegisterExecutableLine("test/pkg", "file.go", 30)
	tracker.RegisterExecutableLine("test/pkg", "file.go", 40)

	// Track some executions
	tracker.TrackExecution("test/pkg", "file.go", 10)
	tracker.TrackExecution("test/pkg", "file.go", 20)
	// Lines 30 and 40 are not executed

	// Generate report
	report := tracker.GenerateReport()

	if report.TotalLines != 4 {
		t.Errorf("Expected 4 total lines, got %d", report.TotalLines)
	}

	if report.CoveredLines != 2 {
		t.Errorf("Expected 2 covered lines, got %d", report.CoveredLines)
	}

	if report.Coverage != 50.0 {
		t.Errorf("Expected 50%% coverage, got %.1f%%", report.Coverage)
	}

	// Check file coverage
	if len(report.Files) != 1 {
		t.Fatalf("Expected 1 file in report, got %d", len(report.Files))
	}

	fc := report.Files[0]
	if fc.Package != "test/pkg" || fc.FileName != "file.go" {
		t.Errorf("Unexpected file info: %s/%s", fc.Package, fc.FileName)
	}

	// Check uncovered lines
	uncovered := fc.GetUncoveredLines()
	if len(uncovered) != 2 {
		t.Errorf("Expected 2 uncovered lines, got %d", len(uncovered))
	}

	// Should contain lines 30 and 40
	if uncovered[0] != 30 || uncovered[1] != 40 {
		t.Errorf("Expected uncovered lines [30, 40], got %v", uncovered)
	}
}

func TestTrackerClear(t *testing.T) {
	tracker := NewTracker()
	tracker.SetEnabled(true)

	// Add some coverage data
	tracker.RegisterExecutableLine("test/pkg", "file.go", 10)
	tracker.TrackExecution("test/pkg", "file.go", 10)

	// Clear coverage data
	tracker.Clear()

	// Coverage data should be empty
	data := tracker.GetCoverageData()
	if len(data) != 0 {
		t.Errorf("Expected empty coverage data after Clear(), got %d packages", len(data))
	}

	// But executable lines should remain
	execLines := tracker.GetExecutableLines()
	if len(execLines) != 1 {
		t.Errorf("Expected executable lines to remain after Clear(), got %d packages", len(execLines))
	}
}
