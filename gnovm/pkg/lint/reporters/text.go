package reporters

import (
	"fmt"
	"io"
)

type TextReporter struct {
	baseReporter
	w io.Writer
}

func NewTextReporter(w io.Writer) *TextReporter {
	return &TextReporter{
		baseReporter: newBaseReporter(),
		w:            w,
	}
}

func (r *TextReporter) Flush() error {
	issues, info, warnings, errors := r.sortAndReset()

	for _, issue := range issues {
		_, _ = fmt.Fprintln(r.w, issue.String())
	}

	total := info + warnings + errors
	if total > 0 {
		_, _ = fmt.Fprintf(r.w, "\nFound %d issue(s): %d error(s), %d warning(s), %d info\n",
			total, errors, warnings, info)
	}

	return nil
}
