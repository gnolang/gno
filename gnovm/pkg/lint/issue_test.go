package lint

import "testing"

func TestIssue_Location(t *testing.T) {
	tests := []struct {
		name     string
		issue    Issue
		expected string
	}{
		{
			name: "basic location",
			issue: Issue{
				Filename: "test.gno",
				Line:     10,
				Column:   5,
			},
			expected: "test.gno:10:5",
		},
		{
			name: "zero line and column",
			issue: Issue{
				Filename: "main.gno",
				Line:     0,
				Column:   0,
			},
			expected: "main.gno:0:0",
		},
		{
			name: "path with directory",
			issue: Issue{
				Filename: "pkg/mypackage/file.gno",
				Line:     42,
				Column:   15,
			},
			expected: "pkg/mypackage/file.gno:42:15",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.issue.Location(); got != tt.expected {
				t.Errorf("Issue.Location() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIssue_String(t *testing.T) {
	tests := []struct {
		name     string
		issue    Issue
		expected string
	}{
		{
			name: "with RuleID",
			issue: Issue{
				RuleID:   "AVL001",
				Severity: SeverityWarning,
				Message:  "unbounded iteration",
				Filename: "test.gno",
				Line:     10,
				Column:   5,
			},
			expected: "test.gno:10:5: warning: unbounded iteration (AVL001)",
		},
		{
			name: "with gnoCode as RuleID",
			issue: Issue{
				RuleID:   "gnoTypeCheckError",
				Severity: SeverityError,
				Message:  "undefined: foo",
				Filename: "main.gno",
				Line:     5,
				Column:   1,
			},
			expected: "main.gno:5:1: error: undefined: foo (gnoTypeCheckError)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.issue.String(); got != tt.expected {
				t.Errorf("Issue.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
