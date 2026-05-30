package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	htmlrenderer "github.com/yuin/goldmark/renderer/html"
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
	// No label attribute → no label strip and no aria-label (no default).
	if strings.Contains(got, "gno-foreign__label") || strings.Contains(got, "aria-label") {
		t.Errorf("unlabeled foreign should have no label strip/aria-label in:\n%s", got)
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

// TestForeign_LinksRenderAsUntrusted locks the gno-foreign link policy:
// links inside the sandbox render as user-generated content — they carry
// rel="noopener nofollow ugc" regardless of type and drop the first-party
// tx/internal trust icons — while a sibling top-level link is unchanged
// (keeps its icon, carries no rel). The hrefs still resolve normally; only
// the trust chrome is stripped. Requires a GnoURL context (else the link
// transformer no-ops) and ExtLinks on the outer instance.
func TestForeign_LinksRenderAsUntrusted(t *testing.T) {
	gnourl, err := weburl.Parse("https://gno.land/r/test")
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	ctxOpts := parser.WithContext(NewGnoParserContext(GnoContext{GnoURL: gnourl}))

	m := goldmark.New()
	ExtForeign.Extend(m, nil)
	ExtColumns.Extend(m)
	ExtAlerts.Extend(m)
	ExtLinks.Extend(m)

	// Top-level internal link (control), then a sandbox with an internal,
	// a help/tx, and an external link.
	src := "[top](/r/foo/hello)\n\n" +
		"<gno-foreign>\n" +
		"[inner](/r/foo/hello) [help](/r/docs/hello$help) [ext](https://example.org/)\n" +
		"</gno-foreign>\n\n"
	var buf bytes.Buffer
	if err := m.Convert([]byte(src), &buf, ctxOpts); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	got := buf.String()

	idx := strings.Index(got, `<div class="gno-foreign__body">`)
	if idx < 0 {
		t.Fatalf("no foreign body div in:\n%s", got)
	}
	top, foreign := got[:idx], got[idx:]

	// Control: top-level internal link keeps its icon and has no rel.
	if !strings.Contains(top, "ico-internal-link") {
		t.Errorf("top-level internal link should keep its icon; got:\n%s", top)
	}
	if strings.Contains(top, "rel=") {
		t.Errorf("top-level internal link must carry no rel; got:\n%s", top)
	}

	// Sandbox: every link carries the ugc rel and resolves normally.
	for _, want := range []string{
		`<a href="/r/foo/hello" rel="noopener nofollow ugc">`,
		`<a href="/r/docs/hello$help" rel="noopener nofollow ugc">`,
		`<a href="https://example.org/" rel="noopener nofollow ugc">`,
	} {
		if !strings.Contains(foreign, want) {
			t.Errorf("foreign link missing %q in:\n%s", want, foreign)
		}
	}

	// Sandbox: first-party trust icons are suppressed; the external-link
	// icon (a "leaves the page" safety hint, not a trust badge) is kept.
	if strings.Contains(foreign, "ico-internal-link") {
		t.Errorf("foreign internal link must not show the internal icon; got:\n%s", foreign)
	}
	if strings.Contains(foreign, "ico-tx-link") {
		t.Errorf("foreign help link must not show the tx icon; got:\n%s", foreign)
	}
	if !strings.Contains(foreign, "ico-external-link") {
		t.Errorf("foreign external link should keep the external icon; got:\n%s", foreign)
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

func TestForeign_CaseVariantCloseEquivalentToLowercase(t *testing.T) {
	// The parser recognizes the close tag case-INSENSITIVELY (goldmark's
	// html.Tokenizer lowercases tag names). Locking this: a case-variant
	// close must produce byte-identical output to the lowercase close —
	// i.e. it terminates the outer block the same way, so "after" renders
	// outside the sandbox in both. The realm-side escaper
	// (foreign.isForeignSentinelLine) mirrors this case-folding; if the
	// parser is ever made case-sensitive without updating the escaper,
	// this test breaks and flags the divergence.
	lower := renderForeignTestCase(t, "\n\n<gno-foreign>\ninside\n</gno-foreign>\nafter\n\n")
	upper := renderForeignTestCase(t, "\n\n<gno-foreign>\ninside\n</GNO-FOREIGN>\nafter\n\n")
	if lower != upper {
		t.Errorf("case-variant close must match lowercase close.\nlower:\n%s\nupper:\n%s", lower, upper)
	}
}

func TestForeign_CaseVariantInnerOpenEquivalentToLowercase(t *testing.T) {
	// A case-variant inner opener must bump the body-framing depth just
	// like a lowercase one (so the matching inner close stays in the body
	// rather than closing the outer block), AND open a nested sandbox at
	// the inner-render level. Byte-equivalence to the all-lowercase
	// nesting proves case-insensitivity holds through BOTH parser passes.
	lower := renderForeignTestCase(t, "\n\n<gno-foreign>\n<gno-foreign>\ninner\n</gno-foreign>\nmid\n</gno-foreign>\n\n")
	upper := renderForeignTestCase(t, "\n\n<gno-foreign>\n<GNO-FOREIGN>\ninner\n</gno-foreign>\nmid\n</gno-foreign>\n\n")
	if lower != upper {
		t.Errorf("case-variant inner open must match lowercase.\nlower:\n%s\nupper:\n%s", lower, upper)
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

func TestForeign_OverCapBodyNotLeakedToOuterHTML(t *testing.T) {
	// REGRESSION: a cap-refused (101st) block must NOT let its opener
	// fall through to the outer raw-HTML block parser — under an UNSAFE
	// outer renderer (gnoweb -html) that would render the unescaped body
	// as live HTML, escaping the sandbox. The refused block must instead
	// be captured opaquely and rendered as a budget marker.
	m := goldmark.New(goldmark.WithRendererOptions(htmlrenderer.WithUnsafe()))
	ExtForeign.Extend(m, nil)
	var b strings.Builder
	for i := 0; i < MaxGnoForeignBlocksPerConvert; i++ {
		b.WriteString("\n\n<gno-foreign>\nok\n</gno-foreign>\n")
	}
	// The (cap+1)th block carries a raw-HTML XSS payload in its body.
	b.WriteString("\n\n<gno-foreign>\n<img src=x onerror=alert(1)>\n</gno-foreign>\n\n")
	var buf bytes.Buffer
	if err := m.Convert([]byte(b.String()), &buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Contains(got, "onerror") {
		t.Errorf("over-cap body leaked as live raw HTML (sandbox escape):\n%s", got)
	}
	if !strings.Contains(got, "render budget exceeded") {
		t.Errorf("expected budget-exceeded marker for the refused block:\n%s", got)
	}
	if n := strings.Count(got, `class="gno-foreign"`); n != MaxGnoForeignBlocksPerConvert {
		t.Errorf("expected exactly %d rendered foreign divs, got %d", MaxGnoForeignBlocksPerConvert, n)
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
	// An empty label= renders no label strip and no aria-label (there is
	// no default label text).
	src := "\n\n<gno-foreign label=\"\">\nbody\n</gno-foreign>\n\n"
	got := renderForeignTestCase(t, src)
	if !strings.Contains(got, `<div class="gno-foreign__body">`) {
		t.Errorf("missing body div in:\n%s", got)
	}
	if strings.Contains(got, "gno-foreign__label") || strings.Contains(got, "aria-label") {
		t.Errorf("empty label should produce no label strip/aria-label in:\n%s", got)
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

func TestForeign_AttributeBearingCloseRecognizedAsSentinel(t *testing.T) {
	// golang.org/x/net/html zeroes the Attr slice on end tags, so the
	// parser cannot tell `</gno-foreign>` from `</gno-foreign attr…>`.
	// This drives the realm-helper escape vector: the helper must
	// mangle attribute-bearing closers, not just bare `</gno-foreign>`.
	// This test pins the parser-side behavior so the helper-side fix
	// has a contract to match against.
	src := "\n\n<gno-foreign>\nfoo\n</gno-foreign label=\"x\">\nAFTER\n\n"
	got := renderForeignTestCase(t, src)
	if !strings.Contains(got, `<div class="gno-foreign"`) {
		t.Fatalf("outer foreign missing: %s", got)
	}
	// "AFTER" must render OUTSIDE the foreign body (the attr-bearing
	// close terminated the outer block).
	bodyStart := strings.Index(got, `<div class="gno-foreign__body">`)
	bodyEnd := strings.Index(got[bodyStart:], "</div>\n</div>")
	if bodyStart < 0 || bodyEnd < 0 {
		t.Fatalf("could not locate foreign body boundaries:\n%s", got)
	}
	bodySlice := got[bodyStart : bodyStart+bodyEnd]
	if strings.Contains(bodySlice, "AFTER") {
		t.Errorf("attr-bearing close did NOT terminate outer; AFTER trapped inside body:\n%s", got)
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
