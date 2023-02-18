package ui

import "testing"

func TestRender(t *testing.T) {
	got := Render("")
	expected := "# UI Demo\n\n[foo](prefixfoo) / [bar](prefixfoo/bar)\n\n\nSimple UI demonstration.\n\n- a text\n- [a relative link](prefixfoobar)\n- another text\n- **a bold text**\n- _italic text_\n- raw markdown with **bold** text in the middle.\n- `some inline code`\n- [a remote link](https://gno.land)\n\nanother string.\n\na paragraph.\n\n\n---\n\n\nI'm the footer.\n\n"

	if got != expected {
		t.Errorf("expected %q, got %q.", expected, got)
	}
}
