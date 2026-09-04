package components_test

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

func TestMarkdownView(t *testing.T) {
	t.Parallel()

	content := []byte("# Title\n\nbody text\n")
	v := components.MarkdownView(content)

	if v.Type != components.MarkdownViewType {
		t.Fatalf("Type = %q, want %q", v.Type, components.MarkdownViewType)
	}

	var buf bytes.Buffer
	if err := v.Render(&buf); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if buf.String() != string(content) {
		t.Fatalf("Render wrote %q, want verbatim %q", buf.String(), string(content))
	}
}
