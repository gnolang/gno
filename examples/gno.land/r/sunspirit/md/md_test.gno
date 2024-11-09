package md

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	output := Render("")
	if !strings.Contains(output, "A simple, flexible, and easy-to-use library for creating markdown documents in gno.land") {
		t.Errorf("invalid output")
	}
}
