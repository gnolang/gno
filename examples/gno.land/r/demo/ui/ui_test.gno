package ui

import (
	"testing"

	"gno.land/p/nt/uassert"
)

func TestRender(t *testing.T) {
	got := Render("")
	expected := "# UI Demo\n\n[foo](r/demo/ui:foo) / [bar](r/demo/ui:foo/bar)\n\n\nSimple UI demonstration.\n\n- a text\n- [a relative link](r/demo/ui:foobar)\n- another text\n- **a bold text**\n- _italic text_\n- raw markdown with **bold** text in the middle.\n- `some inline code`\n- [a remote link](https://gno.land)\n\nanother string.\n\na paragraph.\n\n\n---\n\n\nI'm the footer.\n\n"
	uassert.Equal(t, expected, got)
}
