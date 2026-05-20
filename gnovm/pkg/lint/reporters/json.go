package reporters

import (
	"encoding/json"
	"io"
)

type JSONReporter struct {
	baseReporter
	w io.Writer
}

func NewJSONReporter(w io.Writer) *JSONReporter {
	return &JSONReporter{
		baseReporter: newBaseReporter(),
		w:            w,
	}
}

func (r *JSONReporter) Flush() error {
	issues := r.sortAndReset()

	encoder := json.NewEncoder(r.w)
	encoder.SetIndent("", "\t")
	return encoder.Encode(issues)
}
