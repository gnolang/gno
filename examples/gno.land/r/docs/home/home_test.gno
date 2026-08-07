package home

import (
	"strings"
	"testing"
)

func TestRenderHome(t *testing.T) {
	output := Render("")

	// These strings must appear in the rendered index.
	mustContain := []string{
		"# Gno by Example — r/docs",
		"Official documentation",
		"/r/docs/hello",
		"p/nt/mux/v0",
	}
	for _, want := range mustContain {
		if !strings.Contains(output, want) {
			t.Errorf("Render output missing %q", want)
		}
	}

	// Links to dropped or merged realms must not appear.
	mustNotContain := []string{
		"/r/docs/adder",
		"/r/docs/img_embed",
		"/r/docs/optional_render",
		"/r/docs/charts",
		"/r/docs/safeobjects",
		"/r/docs/avl_pager)",
	}
	for _, bad := range mustNotContain {
		if strings.Contains(output, bad) {
			t.Errorf("Render output contains dead link %q", bad)
		}
	}
}
