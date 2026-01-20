package reporters

import (
	"fmt"
	"io"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

// DirectReporter writes issues to output as soon as they are reported.
// It doesn't buffer, deduplicate or sort results like TextReporter does.
// Used by `gno test` for type-check errors (see gnovm/cmd/gno/test.go).
type DirectReporter struct {
	w        io.Writer
	info     int
	warnings int
	errors   int
}

func NewDirectReporter(w io.Writer) *DirectReporter {
	return &DirectReporter{w: w}
}

func (r *DirectReporter) Report(issue lint.Issue) {
	_, _ = fmt.Fprintln(r.w, issue.String())
	switch issue.Severity {
	case lint.SeverityInfo:
		r.info++
	case lint.SeverityWarning:
		r.warnings++
	case lint.SeverityError:
		r.errors++
	}
}

func (r *DirectReporter) Flush() error {
	return nil
}

func (r *DirectReporter) Summary() (info, warnings, errors int) {
	return r.info, r.warnings, r.errors
}
