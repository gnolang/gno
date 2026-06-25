package hello

import (
	"strings"
	"testing"
)

func TestHello(t *testing.T) {
	expected := "# Hello, 世界！"
	got := Render("")
	if got != expected {
		t.Fatalf("Expected %s, got %s", expected, got)
	}

	got = Render("world")
	expected = "# Hello, world!"
	if got != expected {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

// TestRenderSanitizesPath verifies that markdown-special characters injected via
// path are escaped and cannot break out of the inline heading context.
func TestRenderSanitizesPath(t *testing.T) {
	// An attacker who passes newlines + a blockquote prefix would inject new
	// markdown structure into the rendered page if path were echoed raw.
	injected := "\n\n> injected blockquote"
	got := Render(injected)
	// The output must not contain a bare newline followed by ">",
	// which would be parsed as a blockquote by goldmark.
	if strings.Contains(got, "\n\n>") {
		t.Fatalf("path injection not escaped: got %q", got)
	}
}
