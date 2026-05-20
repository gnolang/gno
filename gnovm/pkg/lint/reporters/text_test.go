package reporters

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

func TestTextReporter_Report(t *testing.T) {
	var buf bytes.Buffer
	r := NewTextReporter(&buf)

	issue := lint.Issue{
		RuleID:   "TEST001",
		Severity: lint.SeverityWarning,
		Message:  "test message",
		Filename: "test.gno",
		Line:     10,
		Column:   5,
	}

	r.Report(issue)

	if len(r.issues) != 1 {
		t.Errorf("issues count = %d, want 1", len(r.issues))
	}
}

func TestTextReporter_Report_Deduplication(t *testing.T) {
	var buf bytes.Buffer
	r := NewTextReporter(&buf)

	issue := lint.Issue{
		RuleID:   "TEST001",
		Severity: lint.SeverityWarning,
		Message:  "test message",
		Filename: "test.gno",
		Line:     10,
		Column:   5,
	}

	// Report same issue twice
	r.Report(issue)
	r.Report(issue)

	if len(r.issues) != 1 {
		t.Errorf("issues count = %d, want 1 (should deduplicate)", len(r.issues))
	}
}

func TestTextReporter_Report_DifferentIssues(t *testing.T) {
	var buf bytes.Buffer
	r := NewTextReporter(&buf)

	issue1 := lint.Issue{
		RuleID:   "TEST001",
		Severity: lint.SeverityWarning,
		Message:  "test message 1",
		Filename: "test.gno",
		Line:     10,
		Column:   5,
	}
	issue2 := lint.Issue{
		RuleID:   "TEST002",
		Severity: lint.SeverityError,
		Message:  "test message 2",
		Filename: "test.gno",
		Line:     20,
		Column:   1,
	}

	r.Report(issue1)
	r.Report(issue2)

	if len(r.issues) != 2 {
		t.Errorf("issues count = %d, want 2", len(r.issues))
	}
}

func TestTextReporter_Report_SeverityCount(t *testing.T) {
	var buf bytes.Buffer
	r := NewTextReporter(&buf)

	r.Report(lint.Issue{RuleID: "T1", Severity: lint.SeverityInfo, Filename: "a.gno", Line: 1, Column: 1})
	r.Report(lint.Issue{RuleID: "T2", Severity: lint.SeverityWarning, Filename: "a.gno", Line: 2, Column: 1})
	r.Report(lint.Issue{RuleID: "T3", Severity: lint.SeverityWarning, Filename: "a.gno", Line: 3, Column: 1})
	r.Report(lint.Issue{RuleID: "T4", Severity: lint.SeverityError, Filename: "a.gno", Line: 4, Column: 1})
	r.Report(lint.Issue{RuleID: "T5", Severity: lint.SeverityError, Filename: "a.gno", Line: 5, Column: 1})
	r.Report(lint.Issue{RuleID: "T6", Severity: lint.SeverityError, Filename: "a.gno", Line: 6, Column: 1})

	if r.info != 1 {
		t.Errorf("info count = %d, want 1", r.info)
	}
	if r.warnings != 2 {
		t.Errorf("warnings count = %d, want 2", r.warnings)
	}
	if r.errors != 3 {
		t.Errorf("errors count = %d, want 3", r.errors)
	}
}

func TestTextReporter_Flush(t *testing.T) {
	var buf bytes.Buffer
	r := NewTextReporter(&buf)

	r.Report(lint.Issue{
		RuleID:   "TEST001",
		Severity: lint.SeverityWarning,
		Message:  "test message",
		Filename: "test.gno",
		Line:     10,
		Column:   5,
	})

	err := r.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

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
	if !strings.Contains(output, "Found 1 issue(s)") {
		t.Error("output should contain summary")
	}
}

func TestTextReporter_Flush_Sorted(t *testing.T) {
	var buf bytes.Buffer
	r := NewTextReporter(&buf)

	// Report in reverse order
	r.Report(lint.Issue{RuleID: "T1", Filename: "z.gno", Line: 20, Column: 1, Severity: lint.SeverityWarning})
	r.Report(lint.Issue{RuleID: "T2", Filename: "a.gno", Line: 10, Column: 1, Severity: lint.SeverityWarning})
	r.Report(lint.Issue{RuleID: "T3", Filename: "a.gno", Line: 5, Column: 1, Severity: lint.SeverityWarning})

	_ = r.Flush()

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Find non-empty lines containing issues
	var issueLines []string
	for _, line := range lines {
		if strings.Contains(line, ".gno:") {
			issueLines = append(issueLines, line)
		}
	}

	if len(issueLines) != 3 {
		t.Fatalf("expected 3 issue lines, got %d", len(issueLines))
	}

	// Should be sorted: a.gno:5, a.gno:10, z.gno:20
	if !strings.Contains(issueLines[0], "a.gno:5") {
		t.Errorf("first issue should be a.gno:5, got %s", issueLines[0])
	}
	if !strings.Contains(issueLines[1], "a.gno:10") {
		t.Errorf("second issue should be a.gno:10, got %s", issueLines[1])
	}
	if !strings.Contains(issueLines[2], "z.gno:20") {
		t.Errorf("third issue should be z.gno:20, got %s", issueLines[2])
	}
}

func TestTextReporter_Flush_NoIssues(t *testing.T) {
	var buf bytes.Buffer
	r := NewTextReporter(&buf)

	err := r.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Found") {
		t.Error("output should not contain summary when no issues")
	}
}
