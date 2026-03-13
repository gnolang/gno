package reporters

import (
	"fmt"
	"sort"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

// baseReporter provides buffered issue collection with deduplication,
// severity counting, and sorted output. Embedded by TextReporter and JSONReporter.
type baseReporter struct {
	issues   []lint.Issue
	seen     map[string]bool
	info     int
	warnings int
	errors   int
}

func newBaseReporter() baseReporter {
	return baseReporter{
		issues: make([]lint.Issue, 0),
		seen:   make(map[string]bool),
	}
}

func (b *baseReporter) Report(issue lint.Issue) {
	key := fmt.Sprintf("%s:%d:%d:%s", issue.Filename, issue.Line, issue.Column, issue.RuleID)
	if b.seen[key] {
		return
	}
	b.seen[key] = true

	b.issues = append(b.issues, issue)

	switch issue.Severity {
	case lint.SeverityInfo:
		b.info++
	case lint.SeverityWarning:
		b.warnings++
	case lint.SeverityError:
		b.errors++
	}
}

func (b *baseReporter) Summary() (info, warnings, errors int) {
	return b.info, b.warnings, b.errors
}

// sortAndReset sorts the buffered issues by filename then line,
// returns them, and resets all state (including severity counts).
// Callers needing counts should call Summary() before sortAndReset().
func (b *baseReporter) sortAndReset() []lint.Issue {
	sort.Slice(b.issues, func(i, j int) bool {
		if b.issues[i].Filename != b.issues[j].Filename {
			return b.issues[i].Filename < b.issues[j].Filename
		}
		return b.issues[i].Line < b.issues[j].Line
	})

	issues := make([]lint.Issue, len(b.issues))
	copy(issues, b.issues)

	b.issues = b.issues[:0]
	clear(b.seen)
	b.info, b.warnings, b.errors = 0, 0, 0

	return issues
}
