package foreign

import (
	"strings"
	"testing"
)

func TestForeign_BasicWrap(t *testing.T) {
	got := Foreign("hello")
	want := "\n\n<gno-foreign>\nhello\n</gno-foreign>\n\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestForeign_LeadingTrailingBlankLines(t *testing.T) {
	// Parser requires a blank line before the opener (CM §4.6). Verify
	// the helper emits exactly that.
	got := Foreign("body")
	if !strings.HasPrefix(got, "\n\n<gno-foreign>\n") {
		t.Errorf("missing blank line before opener: %q", got)
	}
	if !strings.HasSuffix(got, "\n</gno-foreign>\n\n") {
		t.Errorf("missing blank line after closer: %q", got)
	}
}

func TestForeign_NormalizesCRLF(t *testing.T) {
	got := Foreign("line1\r\nline2\r\nline3")
	if strings.Contains(got, "\r") {
		t.Errorf("CRLF not normalized: %q", got)
	}
	if !strings.Contains(got, "line1\nline2\nline3") {
		t.Errorf("body content lost: %q", got)
	}
}

func TestForeign_NormalizesBareCR(t *testing.T) {
	got := Foreign("line1\rline2")
	if strings.Contains(got, "\r") {
		t.Errorf("bare CR not normalized: %q", got)
	}
	if !strings.Contains(got, "line1\nline2") {
		t.Errorf("body content lost: %q", got)
	}
}

func TestForeign_EscapesSentinelClose(t *testing.T) {
	// Unbalanced </gno-foreign> in body would terminate the outer
	// block early. Verify it's escaped to literal text — only the
	// leading "<" is replaced with "&lt;"; the trailing ">" stays
	// literal (rendering as visible ">" in HTML).
	got := Foreign("foo\n</gno-foreign>\nbar")
	if strings.Contains(got, "\n</gno-foreign>\nbar") {
		t.Errorf("unbalanced close not escaped: %q", got)
	}
	if !strings.Contains(got, "&lt;/gno-foreign>") {
		t.Errorf("expected escaped close: %q", got)
	}
}

func TestForeign_EscapesSentinelOpen(t *testing.T) {
	// A bare opener inside body (without matching close in body)
	// would be picked up as a nested inner foreign. Escape it.
	got := Foreign("foo\n<gno-foreign>\nbar")
	if !strings.Contains(got, "\n&lt;gno-foreign>\n") {
		t.Errorf("expected escaped inner opener: %q", got)
	}
}

func TestForeign_EscapesIndentedSentinel(t *testing.T) {
	// Parser strips 0-3 leading spaces before matching; helper
	// must mirror that to catch indented sentinels.
	got := Foreign("foo\n   </gno-foreign>\nbar")
	if !strings.Contains(got, "&lt;/gno-foreign>") {
		t.Errorf("indented sentinel not escaped: %q", got)
	}
}

func TestForeign_EscapesAttributeBearingOpener(t *testing.T) {
	// `<gno-foreign label="x">` IS recognized by the parser as a
	// nested opener. If left unmangled, the parser bumps framingDepth
	// when it sees this line in body bytes, and then consumes the
	// helper's own outer-close as a fake "inner close" — capturing
	// realm content that follows the helper's output INTO the
	// sandbox. Verify the helper escapes the leading "<".
	got := Foreign(`<gno-foreign label="x">`)
	if strings.Contains(got, `<gno-foreign label="x">`) {
		t.Errorf("attribute-bearing opener was NOT mangled — escape vector:\n%q", got)
	}
	if !strings.Contains(got, `&lt;gno-foreign label="x">`) {
		t.Errorf("expected leading '<' escaped: %q", got)
	}
}

func TestForeign_EscapesAttributeBearingCloser(t *testing.T) {
	// golang.org/x/net/html zeroes the Attr slice on end tags, so the
	// parser recognizes ANY `</gno-foreign attr…>` form as a sentinel
	// close. Leaving such a line unmangled in body bytes would close
	// the outer block early (framingDepth=0 case in the parser's
	// Continue), letting body content that follows render OUTSIDE
	// the sandbox at top level. Verify the helper escapes it.
	got := Foreign(`</gno-foreign label="x">`)
	if strings.Contains(got, `> </gno-foreign label="x"> </gno-foreign>`) {
		t.Errorf("attribute-bearing closer was NOT mangled — escape vector:\n%q", got)
	}
	if !strings.Contains(got, `&lt;/gno-foreign label="x">`) {
		t.Errorf("expected leading '<' escaped: %q", got)
	}
}

func TestForeign_EscapesGnoForeignWithExtraContent(t *testing.T) {
	// `<gno-foreign> extra` would not be recognized as a sentinel by
	// the parser (tokenizer yields 2 tokens). But the helper is
	// over-inclusive on purpose: anything that looks like a
	// gno-foreign tag opener or closer gets mangled. Predictable
	// behavior beats narrow matching that diverges from the parser
	// on attribute-bearing forms.
	got := Foreign("<gno-foreign> extra text")
	if !strings.Contains(got, "&lt;gno-foreign> extra text") {
		t.Errorf("expected over-inclusive mangle of `<gno-foreign> extra`: %q", got)
	}
}

func TestForeign_EscapesCaseVariantClose(t *testing.T) {
	// The renderer-side parser recognizes tags case-insensitively
	// (goldmark's tokenizer lowercases tag names). A case-variant close
	// left unescaped would terminate the outer block early and let
	// trailing body render OUTSIDE the sandbox. The helper must escape
	// `</GNO-FOREIGN>` just like `</gno-foreign>`.
	got := Foreign("foo\n</GNO-FOREIGN>\nbar")
	if strings.Contains(got, "\n</GNO-FOREIGN>\nbar") {
		t.Errorf("case-variant close not escaped — sandbox-escape vector: %q", got)
	}
	if !strings.Contains(got, "&lt;/GNO-FOREIGN>") {
		t.Errorf("expected escaped case-variant close: %q", got)
	}
}

func TestForeign_EscapesCaseVariantOpen(t *testing.T) {
	// A case-variant opener left unescaped would be picked up by the
	// parser as a nested inner foreign (framingDepth++), causing the
	// helper's own outer-close to be consumed as a fake inner close and
	// capturing trailing realm content into the sandbox.
	got := Foreign("foo\n<GNO-FOREIGN>\nbar")
	if !strings.Contains(got, "\n&lt;GNO-FOREIGN>\n") {
		t.Errorf("case-variant opener not escaped — capture vector: %q", got)
	}
}

func TestForeign_EscapesMixedCaseSentinel(t *testing.T) {
	// Mixed-case forms (the most likely hand-crafted evasion) must also
	// be caught, including the attribute-bearing mixed-case opener.
	for _, in := range []string{
		"<Gno-Foreign>",
		"</Gno-Foreign>",
		`<GNO-foreign label="x">`,
		"</gNo-fOrEiGn>",
	} {
		got := Foreign(in)
		if !strings.Contains(got, "&lt;") {
			t.Errorf("mixed-case sentinel %q not escaped: %q", in, got)
		}
	}
}

func TestForeignWithLabel_StripsBidiAndZeroWidth(t *testing.T) {
	// A bidi-override (U+202E) or zero-width (U+200C) run in the label
	// is invisible-spoof payload; it must be stripped before the label
	// is spliced into the opener attribute.
	got := ForeignWithLabel("a\u202eb\u200cc", "body")
	if !strings.Contains(got, `label="abc"`) {
		t.Errorf("expected bidi/zero-width-stripped label: %q", got)
	}
}

func TestForeignWithLabel_UnicodeSeparatorsBecomeSpaces(t *testing.T) {
	// U+2028 / U+2029 / U+0085 must not survive in the single-line
	// opener attribute; they fold to spaces like ASCII line breaks.
	got := ForeignWithLabel("a\u2028b\u2029c\u0085d", "body")
	if !strings.Contains(got, `label="a b c d"`) {
		t.Errorf("expected space-folded label: %q", got)
	}
}

func TestForeign_EscapesUnrelatedForeignPrefix(t *testing.T) {
	// `<gno-foreignx>` is a different tag the parser would NOT treat as
	// a sentinel, but the escaper is deliberately over-inclusive (it
	// matches the `<gno-foreign` prefix alone), so it is escaped to
	// visible literal text. Harmless, and the safe side of the trade.
	got := Foreign("<gno-foreignx>")
	if !strings.Contains(got, "&lt;gno-foreignx>") {
		t.Errorf("expected over-inclusive escape of <gno-foreignx>: %q", got)
	}
}

func TestForeign_EscapesSlashAndFormFeedClosers(t *testing.T) {
	// REGRESSION: goldmark's tokenizer ends a tag name at `/` and at
	// form-feed (both HTML tag-name terminators), so `</gno-foreign/>`
	// and `</gno-foreign\f>` are recognized as CLOSERS by the parser. A
	// body line carrying one of these, if it reached the parser
	// unescaped, would close the outer block early — a sandbox escape.
	// The over-inclusive escaper must neutralize every such variant.
	for _, in := range []string{
		"foo\n</gno-foreign/>\nbar",
		"foo\n</gno-foreign/ >\nbar",
		"foo\n</gno-foreign\f>\nbar",
		"foo\n<gno-foreign/>\nbar",
	} {
		got := Foreign(in)
		if !strings.Contains(got, "&lt;") {
			t.Errorf("variant not escaped (sandbox-escape vector): %q -> %q", in, got)
		}
	}
}

func TestForeign_EmptyBody(t *testing.T) {
	got := Foreign("")
	want := "\n\n<gno-foreign>\n\n</gno-foreign>\n\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestForeign_PreservesInternalBlankLines(t *testing.T) {
	got := Foreign("para1\n\npara2")
	if !strings.Contains(got, "para1\n\npara2") {
		t.Errorf("internal blank line lost: %q", got)
	}
}

func TestForeignWithLabel_BasicWrap(t *testing.T) {
	got := ForeignWithLabel("My Label", "body")
	want := "\n\n<gno-foreign label=\"My Label\">\nbody\n</gno-foreign>\n\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestForeignWithLabel_EmptyLabelFallsBack(t *testing.T) {
	// An empty label (or one that becomes empty after sanitization)
	// should be identical to Foreign — no attribute on the opener.
	got := ForeignWithLabel("", "body")
	if strings.Contains(got, "label=") {
		t.Errorf("empty label should not emit attribute: %q", got)
	}
	if got != Foreign("body") {
		t.Errorf("empty label should match Foreign: got %q vs %q", got, Foreign("body"))
	}
}

func TestForeignWithLabel_WhitespaceOnlyLabelFallsBack(t *testing.T) {
	got := ForeignWithLabel("   ", "body")
	if strings.Contains(got, "label=") {
		t.Errorf("whitespace label should not emit attribute: %q", got)
	}
}

func TestForeignWithLabel_EscapesAmpAndQuote(t *testing.T) {
	got := ForeignWithLabel(`Tom & "Jerry"`, "body")
	if !strings.Contains(got, `label="Tom &amp; &quot;Jerry&quot;"`) {
		t.Errorf("expected & and \" escaped: %q", got)
	}
}

func TestForeignWithLabel_EscapesAngleBrackets(t *testing.T) {
	// An unescaped `>` would close the opener tag early. Must be
	// escaped.
	got := ForeignWithLabel(`a<b>c`, "body")
	if !strings.Contains(got, `label="a&lt;b&gt;c"`) {
		t.Errorf("expected < and > escaped: %q", got)
	}
}

func TestForeignWithLabel_StripsNUL(t *testing.T) {
	got := ForeignWithLabel("a\x00b\x00c", "body")
	if strings.Contains(got, "\x00") {
		t.Errorf("NUL not stripped: %q", got)
	}
	if !strings.Contains(got, `label="abc"`) {
		t.Errorf("expected NUL-stripped label: %q", got)
	}
}

func TestForeignWithLabel_ControlCharsBecomeSpaces(t *testing.T) {
	// Newlines and tabs would break the single-line opener required
	// by the parser. They must become spaces.
	got := ForeignWithLabel("line1\nline2\ttab", "body")
	if strings.Contains(got, "\nline2") || strings.Contains(got, "\t") {
		t.Errorf("control chars not converted: %q", got)
	}
	if !strings.Contains(got, "line1 line2 tab") {
		t.Errorf("expected space-separated label: %q", got)
	}
}

func TestForeignWithLabel_TrimsLeadingTrailingSpace(t *testing.T) {
	got := ForeignWithLabel("  hello  ", "body")
	if !strings.Contains(got, `label="hello"`) {
		t.Errorf("expected trimmed label: %q", got)
	}
}

func TestForeignWithLabel_PreservesBodyNormalization(t *testing.T) {
	// Body normalization (CRLF, sentinel escaping) must still apply
	// in the labeled variant.
	got := ForeignWithLabel("L", "a\r\n</gno-foreign>\r\nb")
	if strings.Contains(got, "\r") {
		t.Errorf("CRLF not normalized in labeled variant: %q", got)
	}
	if !strings.Contains(got, "&lt;/gno-foreign>") {
		t.Errorf("sentinel not escaped in labeled variant: %q", got)
	}
}

func TestMaxBlocksPerRender(t *testing.T) {
	// Surfaces gnoweb's per-render foreign-block cap to realms via the
	// chain/markdown native. Pin the contract value; bump here if the
	// renderer cap changes.
	if got := MaxBlocksPerRender(); got != 100 {
		t.Errorf("MaxBlocksPerRender() = %d, want 100", got)
	}
}
