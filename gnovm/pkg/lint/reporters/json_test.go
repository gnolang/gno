package reporters

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

func TestJSONReporter_Report(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSONReporter(&buf)

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

func TestJSONReporter_Report_Deduplication(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSONReporter(&buf)

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

func TestJSONReporter_Report_SeverityCount(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSONReporter(&buf)

	r.Report(lint.Issue{RuleID: "T1", Severity: lint.SeverityInfo, Filename: "a.gno", Line: 1, Column: 1})
	r.Report(lint.Issue{RuleID: "T2", Severity: lint.SeverityWarning, Filename: "a.gno", Line: 2, Column: 1})
	r.Report(lint.Issue{RuleID: "T3", Severity: lint.SeverityError, Filename: "a.gno", Line: 3, Column: 1})

	if r.info != 1 {
		t.Errorf("info count = %d, want 1", r.info)
	}
	if r.warnings != 1 {
		t.Errorf("warnings count = %d, want 1", r.warnings)
	}
	if r.errors != 1 {
		t.Errorf("errors count = %d, want 1", r.errors)
	}
}

func TestJSONReporter_Flush(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSONReporter(&buf)

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

	// Parse the JSON output
	var issues []lint.Issue
	if err := json.Unmarshal(buf.Bytes(), &issues); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(issues) != 1 {
		t.Errorf("parsed %d issues, want 1", len(issues))
	}
	if issues[0].RuleID != "TEST001" {
		t.Errorf("RuleID = %v, want TEST001", issues[0].RuleID)
	}
	if issues[0].Message != "test message" {
		t.Errorf("Message = %v, want 'test message'", issues[0].Message)
	}
}

func TestJSONReporter_Flush_Sorted(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSONReporter(&buf)

	// Report in reverse order
	r.Report(lint.Issue{RuleID: "T1", Filename: "z.gno", Line: 20, Column: 1, Severity: lint.SeverityWarning})
	r.Report(lint.Issue{RuleID: "T2", Filename: "a.gno", Line: 10, Column: 1, Severity: lint.SeverityWarning})
	r.Report(lint.Issue{RuleID: "T3", Filename: "a.gno", Line: 5, Column: 1, Severity: lint.SeverityWarning})

	_ = r.Flush()

	var issues []lint.Issue
	if err := json.Unmarshal(buf.Bytes(), &issues); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(issues) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(issues))
	}

	// Should be sorted: a.gno:5, a.gno:10, z.gno:20
	if issues[0].Filename != "a.gno" || issues[0].Line != 5 {
		t.Errorf("first issue should be a.gno:5, got %s:%d", issues[0].Filename, issues[0].Line)
	}
	if issues[1].Filename != "a.gno" || issues[1].Line != 10 {
		t.Errorf("second issue should be a.gno:10, got %s:%d", issues[1].Filename, issues[1].Line)
	}
	if issues[2].Filename != "z.gno" || issues[2].Line != 20 {
		t.Errorf("third issue should be z.gno:20, got %s:%d", issues[2].Filename, issues[2].Line)
	}
}

func TestJSONReporter_Flush_Empty(t *testing.T) {
	var buf bytes.Buffer
	r := NewJSONReporter(&buf)

	err := r.Flush()
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	var issues []lint.Issue
	if err := json.Unmarshal(buf.Bytes(), &issues); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected empty array, got %d issues", len(issues))
	}
}
