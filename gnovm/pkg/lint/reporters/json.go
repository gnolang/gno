package reporters

import (
	"encoding/json"
	"io"
	"sort"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

type JSONReporter struct {
	w        io.Writer
	issues   []lint.Issue
	info     int
	warnings int
	errors   int
}

func NewJSONReporter(w io.Writer) *JSONReporter {
	return &JSONReporter{
		w:      w,
		issues: make([]lint.Issue, 0),
	}
}

func (r *JSONReporter) Report(issue lint.Issue) {
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
	return encoder.Encode(r.issues)
}

func (r *JSONReporter) Summary() (info, warnings, errors int) {
	return r.info, r.warnings, r.errors
}
