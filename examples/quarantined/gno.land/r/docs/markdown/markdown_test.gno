package markdown

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	output := Render("")
	if !strings.Contains(output, "Gno Flavored Markdown") {
		t.Errorf("invalid output")
	}
}
