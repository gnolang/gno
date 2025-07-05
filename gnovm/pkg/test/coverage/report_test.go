package coverage

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReport_GenerateReport(t *testing.T) {
	tracker := NewTracker()

	// Setup coverage data
	tracker.RegisterExecutableLine("test.gno", 10)
	tracker.RegisterExecutableLine("test.gno", 15)
	tracker.MarkLine("test.gno", 10)

	// Test generating report to file
	tempFile := filepath.Join(t.TempDir(), "coverage.json")
	err := GenerateReport(tracker, tempFile)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Read and verify the report
	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	var report CoverageReport
	err = json.Unmarshal(data, &report)
	if err != nil {
		t.Fatalf("Failed to unmarshal report: %v", err)
	}

	if len(report.Files) != 1 {
		t.Errorf("Expected 1 file in report, got %d", len(report.Files))
	}

	fileCoverage := report.Files["test.gno"]
	if fileCoverage.Total != 2 {
		t.Errorf("Expected 2 total lines, got %d", fileCoverage.Total)
	}
	if fileCoverage.Covered != 1 {
		t.Errorf("Expected 1 covered line, got %d", fileCoverage.Covered)
	}
}

func TestReport_PrintReport(t *testing.T) {
	tracker := NewTracker()
	// Setup coverage data
	tracker.RegisterExecutableLine("test.gno", 10)
	tracker.RegisterExecutableLine("test.gno", 15)
	tracker.MarkLine("test.gno", 10)
	// Capture output
	var buf bytes.Buffer
	err := Print(tracker, &buf)
	if err != nil {
		t.Fatalf("Failed to print report: %v", err)
	}
	output := buf.String()
	t.Logf("PrintReport output: %s", output)
	if !strings.Contains(output, "test.gno") {
		t.Error("Expected filename in output")
	}
	if !strings.Contains(output, "%") {
		t.Error("Expected coverage percentage in output")
	}
}
