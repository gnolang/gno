package markdown

import (
	"fmt"
	"strings"
	"testing"
)

// TestFormSelectsLinear renders a form holding many selects, each under its
// own name. Rendering one select scanned every element of the form for its
// options, so n names cost n*n (20,000 took 1.5 s).
func TestFormSelectsLinear(t *testing.T) {
	assertLinear(t, func(n int) string {
		var b strings.Builder
		b.WriteString("<gno-form>\n")
		for i := range n {
			fmt.Fprintf(&b, "<gno-select name=\"s%d\" value=\"v\" />\n", i)
		}
		b.WriteString("</gno-form>\n")
		return b.String()
	}, 5000)
}
