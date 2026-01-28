package reporters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

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

func TestDirectReporter_Report_SeverityCount(t *testing.T) {
	var buf bytes.Buffer
	r := NewDirectReporter(&buf)

	// Report issues of each severity
	r.Report(lint.Issue{RuleID: "T1", Severity: lint.SeverityInfo, Filename: "a.gno", Line: 1, Column: 1})
	r.Report(lint.Issue{RuleID: "T2", Severity: lint.SeverityWarning, Filename: "a.gno", Line: 2, Column: 1})
	r.Report(lint.Issue{RuleID: "T3", Severity: lint.SeverityWarning, Filename: "a.gno", Line: 3, Column: 1})
	r.Report(lint.Issue{RuleID: "T4", Severity: lint.SeverityError, Filename: "a.gno", Line: 4, Column: 1})
	r.Report(lint.Issue{RuleID: "T5", Severity: lint.SeverityError, Filename: "a.gno", Line: 5, Column: 1})

	info, warnings, errors := r.Summary()
	if info != 1 {
		t.Errorf("info = %d, want 1", info)
	}
	if warnings != 2 {
		t.Errorf("warnings = %d, want 2", warnings)
	}
	if errors != 2 {
		t.Errorf("errors = %d, want 2", errors)
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
