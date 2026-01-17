package reporters

import (
	"fmt"
	"io"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

// DirectReporter prints issues immediately without buffering, for real-time output.
type DirectReporter struct {
	w      io.Writer
	errors int
}

func NewDirectReporter(w io.Writer) *DirectReporter {
	return &DirectReporter{w: w}
}

func (r *DirectReporter) Report(issue lint.Issue) {
	_, _ = fmt.Fprintln(r.w, issue.String())
	if issue.Severity == lint.SeverityError {
		r.errors++
	}
}

func (r *DirectReporter) Flush() error {
	return nil
}

func (r *DirectReporter) Summary() (info, warnings, errors int) {
	return 0, 0, r.errors
}
