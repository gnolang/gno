package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

// buildTestMarkdownWithForeign returns a goldmark instance with the
// foreign extension (and the structural family that participates in
// the nesting cap) registered. Image validator is nil for tests.
func buildTestMarkdownWithForeign() goldmark.Markdown {
	m := goldmark.New()
	ExtForeign.Extend(m, nil)
	ExtColumns.Extend(m)
	ExtAlerts.Extend(m)
	return m
}

// renderForeignTestCase converts the given markdown via the test
// markdown instance and returns the rendered HTML.
func renderForeignTestCase(t *testing.T, src string) string {
	t.Helper()
	m := buildTestMarkdownWithForeign()
	var buf bytes.Buffer
	if err := m.Convert([]byte(src), &buf); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	return buf.String()
}

func TestForeign_EmptyBody(t *testing.T) {
	src := "\n\n<gno-foreign>\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	if !strings.Contains(got, `<div class="gno-foreign"`) {
		t.Errorf("missing outer foreign div in:\n%s", got)
	}
	if !strings.Contains(got, `<div class="gno-foreign__body">`) {
		t.Errorf("missing body div in:\n%s", got)
	}
	if !strings.Contains(got, `aria-label="external content"`) {
		t.Errorf("missing default aria-label in:\n%s", got)
	}
}

func TestForeign_SimpleBody(t *testing.T) {
	src := "\n\n<gno-foreign>\nHello world.\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	if !strings.Contains(got, `<div class="gno-foreign__body">`) {
		t.Fatalf("no body div in:\n%s", got)
	}
	if !strings.Contains(got, "Hello world.") {
		t.Errorf("body text missing in:\n%s", got)
	}
}

func TestForeign_LiteralCloseInBody_DoesNotEscape(t *testing.T) {
	// Outer parser depth-counts inner <gno-foreign>/</gno-foreign>
	// pairs at byte level. A close inside a paired-up inner block
	// should NOT terminate the outer.
	src := "\n\n<gno-foreign>\n<gno-foreign>\ninner body\n</gno-foreign>\nouter still open\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	// Outer should contain both "inner body" (from the nested foreign's
	// own rendered output) and "outer still open" (from the outer
	// body bytes after the inner closes).
	if !strings.Contains(got, "inner body") {
		t.Errorf("inner body missing in:\n%s", got)
	}
	if !strings.Contains(got, "outer still open") {
		t.Errorf("outer-still-open missing in:\n%s", got)
	}
}

func TestForeign_NestedDepth2(t *testing.T) {
	src := "\n\n<gno-foreign>\n<gno-foreign>\nnested\n</gno-foreign>\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	// Expect two opening foreign divs in the rendered HTML.
	openCount := strings.Count(got, `<div class="gno-foreign"`)
	if openCount != 2 {
		t.Errorf("expected 2 foreign divs, got %d in:\n%s", openCount, got)
	}
	if !strings.Contains(got, "nested") {
		t.Errorf("nested body missing in:\n%s", got)
	}
}

func TestForeign_DepthCapAt5_FifthRefused(t *testing.T) {
	// 5 nested <gno-foreign> openers — the 5th must be rejected
	// (cap is 4 across the gno-* family).
	var b strings.Builder
	b.WriteString("\n\n")
	for i := 0; i < 5; i++ {
		b.WriteString("<gno-foreign>\n")
	}
	b.WriteString("deepest body\n")
	for i := 0; i < 5; i++ {
		b.WriteString("</gno-foreign>\n")
	}
	b.WriteString("\n")
	got := renderForeignTestCase(t, b.String())
	// Expect 4 successful opening foreign divs and the 5th refused.
	openCount := strings.Count(got, `<div class="gno-foreign"`)
	if openCount != 4 {
		t.Errorf("expected 4 successfully-opened foreigns (cap=4), got %d in:\n%s", openCount, got)
	}
}

func TestForeign_OpenerWithLabel(t *testing.T) {
	// `<gno-foreign label="My Label">` is recognized; the label
	// flows through to the rendered aria-label and label div.
	src := "\n\n<gno-foreign label=\"My Label\">\nbody\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	if !strings.Contains(got, `aria-label="My Label"`) {
		t.Errorf("missing custom aria-label in:\n%s", got)
	}
	if !strings.Contains(got, `>My Label<`) {
		t.Errorf("missing label text in:\n%s", got)
	}
	if !strings.Contains(got, "body") {
		t.Errorf("body content missing in:\n%s", got)
	}
}

func TestForeign_OpenerWithEmptyLabel(t *testing.T) {
	// Empty label falls back to the default "external content".
	src := "\n\n<gno-foreign label=\"\">\nbody\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	if !strings.Contains(got, `aria-label="external content"`) {
		t.Errorf("missing default aria-label fallback in:\n%s", got)
	}
}

func TestForeign_OpenerWithUnknownAttribute_FallsThrough(t *testing.T) {
	// Only `label` is recognized. Any other attribute makes the
	// opener fail closed (fall through to raw HTML).
	src := "\n\n<gno-foreign foo=\"x\">\nbody\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	if strings.Contains(got, `<div class="gno-foreign"`) {
		t.Errorf("unknown attribute should NOT create a foreign div; got:\n%s", got)
	}
}

func TestForeign_OpenerWithMultipleAttributes_FallsThrough(t *testing.T) {
	// Multiple attributes are rejected even when `label` is one of
	// them — keeps the recognizer strict.
	src := "\n\n<gno-foreign label=\"a\" foo=\"b\">\nbody\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	if strings.Contains(got, `<div class="gno-foreign"`) {
		t.Errorf("multi-attribute opener should NOT create a foreign div; got:\n%s", got)
	}
}

func TestForeign_LabelHTMLSpecialCharsEscaped(t *testing.T) {
	// HTML entities in the attribute decode to special characters in
	// the parsed label. The renderer must HTML-escape them back so
	// the rendered HTML stays well-formed.
	src := "\n\n<gno-foreign label=\"&lt;x&amp;y&gt;\">\nbody\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	if strings.Contains(got, "<x") || strings.Contains(got, "y>") {
		t.Errorf("special chars not re-escaped in render:\n%s", got)
	}
	if !strings.Contains(got, "&lt;x&amp;y&gt;") {
		t.Errorf("expected escaped label in output:\n%s", got)
	}
}

func TestForeign_NoBlankLineBeforeOpener_FallsThrough(t *testing.T) {
	// Per CM §4.6, Type-7 HTML blocks (which gno-foreign matches)
	// cannot interrupt a paragraph. Without the blank line before
	// the opener, our parser's CanInterruptParagraph=false makes
	// the line absorbed into the preceding paragraph rather than
	// opening a foreign block.
	src := "paragraph text\n<gno-foreign>\nbody\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	if strings.Contains(got, `<div class="gno-foreign"`) {
		t.Errorf("opener without blank-line-before should NOT create foreign div; got:\n%s", got)
	}
}

func TestForeign_UnclosedSynthCloses(t *testing.T) {
	// Opener with no matching close — AST transformer must synth-
	// close so the rendered output stays well-balanced.
	src := "\n\n<gno-foreign>\nunclosed body\n"
	got := renderForeignTestCase(t, src)
	if !strings.Contains(got, `<div class="gno-foreign"`) {
		t.Fatalf("unclosed foreign missing outer div:\n%s", got)
	}
	if !strings.Contains(got, "unclosed body") {
		t.Errorf("body content missing:\n%s", got)
	}
	if strings.Count(got, "</div>") < 2 {
		t.Errorf("expected at least two </div> closures for outer+body, got:\n%s", got)
	}
}
