package reporters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

func TestNewDirectReporter(t *testing.T) {
	var buf bytes.Buffer
	r := NewDirectReporter(&buf)

	if r == nil {
		t.Fatal("NewDirectReporter() returned nil")
	}
	if r.w != &buf {
		t.Error("writer not set correctly")
	}
}

func TestDirectReporter_Report(t *testing.T) {
	var buf bytes.Buffer
	r := NewDirectReporter(&buf)

	issue := lint.Issue{
		RuleID:   "TEST001",
		Severity: lint.SeverityWarning,
		Message:  "test message",
		Filename: "test.gno",
		Line:     10,
		Column:   5,
	}

	r.Report(issue)

	output := buf.String()
	if !strings.Contains(output, "test.gno:10:5") {
		t.Error("output should contain issue location")
	}
	if !strings.Contains(output, "TEST001") {
		t.Error("output should contain rule ID")
	}
	if !strings.Contains(output, "test message") {
		t.Error("output should contain message")
	}
}

func TestDirectReporter_Report_ImmediateOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewDirectReporter(&buf)

	// Report first issue
	r.Report(lint.Issue{
		RuleID:   "T1",
		Severity: lint.SeverityWarning,
		Message:  "first",
		Filename: "a.gno",
		Line:     1,
		Column:   1,
	})

	// Output should be available immediately (no buffering)
	if !strings.Contains(buf.String(), "first") {
		t.Error("first issue should be output immediately")
	}

	// Report second issue
	r.Report(lint.Issue{
		RuleID:   "T2",
		Severity: lint.SeverityError,
		Message:  "second",
		Filename: "b.gno",
		Line:     1,
		Column:   1,
	})

	output := buf.String()
	if !strings.Contains(output, "second") {
		t.Error("second issue should be output immediately")
	}
}

func TestDirectReporter_Report_ErrorCount(t *testing.T) {
	var buf bytes.Buffer
	r := NewDirectReporter(&buf)

	// Info and warning don't increment error count
	r.Report(lint.Issue{RuleID: "T1", Severity: lint.SeverityInfo, Filename: "a.gno", Line: 1, Column: 1})
	r.Report(lint.Issue{RuleID: "T2", Severity: lint.SeverityWarning, Filename: "a.gno", Line: 2, Column: 1})

	if r.errors != 0 {
		t.Errorf("errors = %d, want 0 after info and warning", r.errors)
	}

	// Error increments error count
	r.Report(lint.Issue{RuleID: "T3", Severity: lint.SeverityError, Filename: "a.gno", Line: 3, Column: 1})
	r.Report(lint.Issue{RuleID: "T4", Severity: lint.SeverityError, Filename: "a.gno", Line: 4, Column: 1})

	if r.errors != 2 {
		t.Errorf("errors = %d, want 2", r.errors)
	}
}

func TestDirectReporter_Flush(t *testing.T) {
	var buf bytes.Buffer
	r := NewDirectReporter(&buf)

	r.Report(lint.Issue{
		RuleID:   "TEST001",
		Severity: lint.SeverityWarning,
		Message:  "test message",
		Filename: "test.gno",
		Line:     10,
		Column:   5,
	})

	// Flush should do nothing and return nil
	err := r.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}
}

func TestDirectReporter_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewDirectReporter(&buf)

	r.Report(lint.Issue{RuleID: "T1", Severity: lint.SeverityInfo, Filename: "a.gno", Line: 1, Column: 1})
	r.Report(lint.Issue{RuleID: "T2", Severity: lint.SeverityWarning, Filename: "a.gno", Line: 2, Column: 1})
	r.Report(lint.Issue{RuleID: "T3", Severity: lint.SeverityError, Filename: "a.gno", Line: 3, Column: 1})

	info, warnings, errors := r.Summary()

	// DirectReporter only tracks errors
	if info != 0 {
		t.Errorf("info = %d, want 0 (not tracked)", info)
	}
	if warnings != 0 {
		t.Errorf("warnings = %d, want 0 (not tracked)", warnings)
	}
	if errors != 1 {
		t.Errorf("errors = %d, want 1", errors)
	}
}

func TestDirectReporter_NoDeduplication(t *testing.T) {
	var buf bytes.Buffer
	r := NewDirectReporter(&buf)

	issue := lint.Issue{
		RuleID:   "TEST001",
		Severity: lint.SeverityWarning,
		Message:  "test message",
		Filename: "test.gno",
		Line:     10,
		Column:   5,
	}

	// Report same issue twice - DirectReporter does NOT deduplicate
	r.Report(issue)
	r.Report(issue)

	output := buf.String()
	count := strings.Count(output, "TEST001")
	if count != 2 {
		t.Errorf("TEST001 appeared %d times, want 2 (no deduplication)", count)
	}
}
