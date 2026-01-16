package reporters

import (
	"fmt"
	"io"
	"sort"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

type TextReporter struct {
	w        io.Writer
	issues   []lint.Issue
	info     int
	warnings int
	errors   int
}

func NewTextReporter(w io.Writer) *TextReporter {
	return &TextReporter{
		w:      w,
		issues: make([]lint.Issue, 0),
	}
}

func (r *TextReporter) Report(issue lint.Issue) {
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

func (r *TextReporter) Flush() error {
	sort.Slice(r.issues, func(i, j int) bool {
		if r.issues[i].Filename != r.issues[j].Filename {
			return r.issues[i].Filename < r.issues[j].Filename
		}
		return r.issues[i].Line < r.issues[j].Line
	})

	for _, issue := range r.issues {
		fmt.Fprintln(r.w, issue.String())
	}

	total := r.info + r.warnings + r.errors
	if total > 0 {
		fmt.Fprintf(r.w, "\nFound %d issue(s): %d error(s), %d warning(s), %d info\n",
			total, r.errors, r.warnings, r.info)
	}

	return nil
}

func (r *TextReporter) Summary() (info, warnings, errors int) {
	return r.info, r.warnings, r.errors
}
