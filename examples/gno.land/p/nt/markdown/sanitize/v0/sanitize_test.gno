package sanitize

import "testing"

func TestStripBidiAndZeroWidthRe(t *testing.T) {
	// Re-export sanity check.
	if got := StripBidiAndZeroWidth("a\u200Bb"); got != "ab" {
		t.Errorf("re-export StripBidiAndZeroWidth failed: got %q", got)
	}
}

func TestNormalizeBreaksRe(t *testing.T) {
	if got := NormalizeBreaks("a\r\nb"); got != "a\nb" {
		t.Errorf("re-export NormalizeBreaks failed: got %q", got)
	}
}

func TestInlineText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"a*b", `a\*b`},
		{"a|b", "a|b"},                          // | NOT escaped
		{"a=b", "a=b"},                          // = NOT escaped
		{"line1\nline2", "line1 line2"},          // newline folded
		{"line1\r\nline2", "line1 line2"},        // CRLF normalized then folded
		{"a\u2028b", "a b"},                     // U+2028 folded
		{"a\u0085b", "a b"},                     // NEL folded
		{"a\u200Bb", "ab"},                      // ZWSP stripped
		{"hello *world*", `hello \*world\*`},
	}
	for _, c := range cases {
		if got := InlineText(c.in); got != c.want {
			t.Errorf("InlineText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBlock(t *testing.T) {
	cases := []struct{ in, want string }{
		{"hello world\n", "hello world\n"},
		{"# heading\n", "\\# heading\n"},
		{"> quote\n", "\\> quote\n"},
		{"```\ncode\n", "```\ncode\n```\n"}, // fence auto-close at EOF
	}
	for _, c := range cases {
		if got := Block(c.in); got != c.want {
			t.Errorf("Block(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLinkTitle(t *testing.T) {
	cases := []struct{ in, want string }{
		{`he said "hi"`, `he said \"hi\"`},
		{`it's nice`, `it\'s nice`},
		{"line1\nline2", "line1 line2"},
	}
	for _, c := range cases {
		if got := LinkTitle(c.in); got != c.want {
			t.Errorf("LinkTitle(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTableCell(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain cell", "plain cell"},
		{"a|b", `a\|b`},
		{"a\tb", "a b"},
		{"a*b|c", `a\*b\|c`},
	}
	for _, c := range cases {
		if got := TableCell(c.in); got != c.want {
			t.Errorf("TableCell(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestHTMLEscape(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"<script>", "&lt;script&gt;"},
		{`a & b`, "a &amp; b"},
		{`"quoted"`, "&#34;quoted&#34;"},
	}
	for _, c := range cases {
		if got := HTMLEscape(c.in); got != c.want {
			t.Errorf("HTMLEscape(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://example.com/x", "https://example.com/x"},
		{"http://example.com", "http://example.com"},
		{"mailto:a@b.com", "mailto:a@b.com"},
		{"mailto:a@b.com?body=phish", ""}, // prefill phishing rejected
		{"//evil.com", ""},                 // protocol-relative rejected
		{"javascript:alert(1)", ""},        // bad scheme
		{"/r/foo", "/r/foo"},               // relative
		{"./local", "./local"},
		{"#section", "#section"},           // fragment-only
		{"", ""},
		{"   ", ""},
		{"https://a.com/path with space", "https://a.com/path%20with%20space"},
	}
	for _, c := range cases {
		if got := URL(c.in); got != c.want {
			t.Errorf("URL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestImageURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://example.com/img.png", "https://example.com/img.png"},
		{"mailto:a@b.com", ""},          // mailto rejected for images
		{"data:image/svg+xml,<svg/>", "data:image/svg+xml,%3Csvg/%3E"},
		{"data:image/png;base64,XXX", "data:image/png;base64,XXX"},
		{"data:text/html,<script>", ""}, // bad data subset
		{"javascript:alert(1)", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := ImageURL(c.in); got != c.want {
			t.Errorf("ImageURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestUserName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"alice", "alice"},
		{"alice123", "alice123"},
		{"alice_bob-cat", "alice_bob-cat"},
		{"Alice", ""},  // uppercase first
		{"1alice", ""}, // digit first
		{"", ""},
		{"a\u200Blice", "alice"}, // bidi stripped, then matches
	}
	for _, c := range cases {
		if got := UserName(c.in); got != c.want {
			t.Errorf("UserName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBechString(t *testing.T) {
	addrG := "g1abc123def456ghi789jkl012mno345p"
	cases := []struct {
		s, prefix string
		want      string
	}{
		{addrG, "g", addrG},
		{addrG, "", addrG},     // any prefix
		{addrG, "gpub", ""},   // wrong prefix
		{"gpub1abc123def456ghijklmn", "gpub", "gpub1abc123def456ghijklmn"},
		{"gpub1abc123def456ghijklmn", "", "gpub1abc123def456ghijklmn"},
		{"b1xyz789abc123def456", "", "b1xyz789abc123def456"}, // any-prefix mode allows b1...
		{"g1ABC", "g", ""},     // uppercase rejected
		{"x", "g", ""},
		{"", "g", ""},
	}
	for _, c := range cases {
		if got := BechString(c.s, c.prefix); got != c.want {
			t.Errorf("BechString(%q,%q) = %q, want %q", c.s, c.prefix, got, c.want)
		}
	}
}

func TestFootnoteRef(t *testing.T) {
	cases := []struct{ in, want string }{
		{"note1", "note1"},
		{"Note_A-1", "Note_A-1"},
		{"with space", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := FootnoteRef(c.in); got != c.want {
			t.Errorf("FootnoteRef(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLanguageName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"go", "go"},
		{"c++", "c++"},
		{"python3", "python3"},
		{"objective-c", "objective-c"},
		{"with space", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := LanguageName(c.in); got != c.want {
			t.Errorf("LanguageName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNestedPrefix(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"  ", "  "},
		{"\t", "\t"},
		{"> ", "> "},
		{"> > ", "> > "},
		{"## ", ""}, // markdown-active prefix rejected
		{"- ", ""},
	}
	for _, c := range cases {
		if got := NestedPrefix(c.in); got != c.want {
			t.Errorf("NestedPrefix(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCodeFence(t *testing.T) {
	if got := CodeFence("```", 3); got != "````" {
		t.Errorf("CodeFence: got %q, want %q", got, "````")
	}
	if got := CodeFence("", 3); got != "```" {
		t.Errorf("CodeFence empty: got %q, want %q", got, "```")
	}
}
