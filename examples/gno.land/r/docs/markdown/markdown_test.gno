package markdown

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	output := Render("")
	if !strings.Contains(output, "\nMarkdown Quick Reference\n") {
		t.Errorf("invalid output")
	}
}
