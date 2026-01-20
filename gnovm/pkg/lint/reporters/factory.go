package reporters

import (
	"fmt"
	"io"

	"github.com/gnolang/gno/gnovm/pkg/lint"
)

const (
	FormatText = "text"
	FormatJSON = "json"
)

func NewReporter(format string, w io.Writer) (lint.Reporter, error) {
	switch format {
	case FormatText, "":
		return NewTextReporter(w), nil
	case FormatJSON:
		return NewJSONReporter(w), nil
	default:
		return nil, fmt.Errorf("unknown output format: %q (available: text, json)", format)
	}
}
