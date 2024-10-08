package md

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	output := Render("")
	if !strings.Contains(output, "Markdown rendering") {
		t.Errorf("invalid output")
	}
}
