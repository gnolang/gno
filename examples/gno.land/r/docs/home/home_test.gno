package home

import (
	"strings"
	"testing"
)

func TestRenderHome(t *testing.T) {
	output := Render("")

	// Check for the presence of key sections
	if !contains(output, "# Gno Examples Documentation") {
		t.Errorf("Render output is missing the title.")
	}
	if !contains(output, "Official documentation") {
		t.Errorf("Render output is missing the official documentation link.")
	}
}

func contains(s, substr string) bool {
	return strings.Index(s, substr) >= 0
}
