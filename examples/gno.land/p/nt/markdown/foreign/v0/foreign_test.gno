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
	// block early. Verify it's escaped to literal text.
	got := Foreign("foo\n</gno-foreign>\nbar")
	if strings.Contains(got, "\n</gno-foreign>\nbar") {
		t.Errorf("unbalanced close not escaped: %q", got)
	}
	if !strings.Contains(got, "&lt;/gno-foreign&gt;") {
		t.Errorf("expected escaped close: %q", got)
	}
}

func TestForeign_EscapesSentinelOpen(t *testing.T) {
	// A bare opener inside body (without matching close in body)
	// would be picked up as a nested inner foreign. Escape it.
	got := Foreign("foo\n<gno-foreign>\nbar")
	// Outer opener stays literal (first line after the leading
	// blanks); only the inner sentinel match gets escaped.
	if !strings.Contains(got, "\n&lt;gno-foreign&gt;\n") {
		t.Errorf("expected escaped inner opener: %q", got)
	}
}

func TestForeign_EscapesIndentedSentinel(t *testing.T) {
	// Parser strips 0-3 leading spaces before matching; helper
	// must mirror that to catch indented sentinels.
	got := Foreign("foo\n   </gno-foreign>\nbar")
	if !strings.Contains(got, "&lt;/gno-foreign&gt;") {
		t.Errorf("indented sentinel not escaped: %q", got)
	}
}

func TestForeign_DoesNotEscapeAttributeBearing(t *testing.T) {
	// `<gno-foreign label="x">` is not a sentinel match (parser
	// rejects attribute-bearing tags). Body should be untouched.
	body := `<gno-foreign label="x">`
	got := Foreign(body)
	if !strings.Contains(got, body) {
		t.Errorf("attribute-bearing tag was modified: %q", got)
	}
}

func TestForeign_DoesNotEscapePartialMatch(t *testing.T) {
	// `<gno-foreign> extra` has trailing content; tokenizer yields
	// more than one token so it's not a sentinel match. Leave it.
	body := "<gno-foreign> extra text"
	got := Foreign(body)
	if !strings.Contains(got, body) {
		t.Errorf("partial-match line was modified: %q", got)
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
	if !strings.Contains(got, "&lt;/gno-foreign&gt;") {
		t.Errorf("sentinel not escaped in labeled variant: %q", got)
	}
}
