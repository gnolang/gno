package reporters

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

type JSONReporter struct {
	w        io.Writer
	issues   []lint.Issue
	seen     map[string]bool
	info     int
	warnings int
	errors   int
}

func NewJSONReporter(w io.Writer) *JSONReporter {
	return &JSONReporter{
		w:      w,
		issues: make([]lint.Issue, 0),
		seen:   make(map[string]bool),
	}
}

func (r *JSONReporter) Report(issue lint.Issue) {
	key := fmt.Sprintf("%s:%d:%d:%s", issue.Filename, issue.Line, issue.Column, issue.RuleID)
	if r.seen[key] {
		return
	}
	r.seen[key] = true

	r.issues = append(r.issues, issue)

	switch issue.Severity {
	case lint.SeverityInfo:
		r.info++
	case lint.SeverityWarning:
		r.warnings++
	case lint.SeverityError:
		r.errors++
	}
}

func (r *JSONReporter) Flush() error {
	sort.Slice(r.issues, func(i, j int) bool {
		if r.issues[i].Filename != r.issues[j].Filename {
			return r.issues[i].Filename < r.issues[j].Filename
		}
		return r.issues[i].Line < r.issues[j].Line
	})

	encoder := json.NewEncoder(r.w)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(r.issues)

	r.issues = r.issues[:0]
	clear(r.seen)
	r.info, r.warnings, r.errors = 0, 0, 0

	return err
}

func (r *JSONReporter) Summary() (info, warnings, errors int) {
	return r.info, r.warnings, r.errors
}
