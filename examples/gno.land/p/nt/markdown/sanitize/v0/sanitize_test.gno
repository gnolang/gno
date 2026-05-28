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
		{"a|b", "a|b"},                    // | NOT escaped
		{"a=b", "a=b"},                    // = NOT escaped
		{"line1\nline2", "line1 line2"},   // newline folded
		{"line1\r\nline2", "line1 line2"}, // CRLF normalized then folded
		{"a\u2028b", "a b"},               // U+2028 folded
		{"a\u0085b", "a b"},               // NEL folded
		{"a\u200Bb", "ab"},                // ZWSP stripped
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
		got := Block(c.in)
		if got != c.want {
			t.Errorf("Block(%q) = %q, want %q", c.in, got, c.want)
		}
		// Idempotency: Block(Block(in)) must be byte-identical to
		// Block(in) for every input in the table.
		twice := Block(got)
		if twice != got {
			t.Errorf("Block not idempotent for %q: Block(once)=%q, Block(twice)=%q", c.in, got, twice)
		}
	}
}

func TestBlockRich(t *testing.T) {
	// BlockRich prepends "\n" to its output as part of the cross-
	// boundary setext defense. Every expected output here includes
	// that leading "\n".
	cases := []struct{ in, want string }{
		// Block-level markdown that Block escapes — preserved by BlockRich.
		{"hello world\n", "\nhello world\n"},
		{"# heading\n", "\n# heading\n"},
		{"> quote\n", "\n> quote\n"},
		{"- item\n", "\n- item\n"},
		{"1. item\n", "\n1. item\n"},
		{"---\n", "\n\\---\n"},           // first-line setext-h2 → escaped by qualifying-setext pre-pass
		{"***\n", "\n***\n"},             // thematic break NOT setext-h2 — preserved
		{"text\n===\n", "\ntext\n===\n"}, // deeper setext preserved (user-authored above)
		// Realm-binding defenses STILL ON.
		{"<gno-card>\n", "\n\\<gno-card>\n"},
		{"[t][l]\n", "\n\\[t\\]\\[l\\]\n"},
		{"[^name]\n", "\n\\[^name\\]\n"},
		{"[x]: y\n", "\n"}, // LRD stripped → BlockRich returns just the leading "\n"
		{"```\ncode\n", "\n```\ncode\n```\n"},
	}
	for _, c := range cases {
		got := BlockRich(c.in)
		if got != c.want {
			t.Errorf("BlockRich(%q) = %q, want %q", c.in, got, c.want)
		}
		// Idempotency: BlockRich(BlockRich(in)) must be byte-identical
		// to BlockRich(in) for every input in the table.
		twice := BlockRich(got)
		if twice != got {
			t.Errorf("BlockRich not idempotent for %q: BlockRich(once)=%q, BlockRich(twice)=%q", c.in, got, twice)
		}
	}
}

func TestNeuterLeadingSetextIfQualifying(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"leading-setext-h1", "===\nbody\n", "\\===\nbody\n"},
		{"leading-setext-h2", "---\nbody\n", "\\---\nbody\n"},
		{"leading-setext-indented", "   ===\nbody\n", "   \\===\nbody\n"},
		{"leading-setext-indented-4", "    ===\nbody\n", "    ===\nbody\n"}, // indented code
		{"leading-setext-trailing-ws", "===   \nbody\n", "\\===   \nbody\n"},
		{"leading-setext-mixed-chars", "=-=-\nbody\n", "=-=-\nbody\n"},         // mixed; not setext
		{"leading-setext-has-content", "=== text\nbody\n", "=== text\nbody\n"}, // mixed content
		{"setext-deeper-untouched", "title\n===\nfoo\n", "title\n===\nfoo\n"},
		{"leading-thematic-asterisk", "***\nbody\n", "***\nbody\n"}, // not setext, untouched
		{"leading-thematic-underscore", "___\nbody\n", "___\nbody\n"},
		{"leading-blank-then-setext", "\n===\nbody\n", "\n\\===\nbody\n"},
		{"leading-blank-ws-then-setext", "   \n===\nbody\n", "   \n\\===\nbody\n"},
		{"text-first", "hello\n===\n", "hello\n===\n"}, // text before === — user-authored
		{"empty", "", ""},
		{"all-blank", "   \n\n", "   \n\n"},
		{"only-equals", "===", "\\==="},
	}
	for _, c := range cases {
		if got := neuterLeadingSetextIfQualifying(c.in); got != c.want {
			t.Errorf("%s: neuterLeadingSetextIfQualifying(%q) = %q, want %q", c.name, c.in, got, c.want)
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
		{"mailto:a@b.com?body=phish", ""},                  // prefill phishing rejected
		{"mailto:a@b.com?BODY=phish", ""},                  // RFC 6068 §2: hfname case-insensitive
		{"mailto:a@b.com?Body=phish", ""},                  // mixed case
		{"mailto:a@b.com?subject=hi&body=phish", ""},       // body as second param
		{"mailto:a@b.com?subject=hi&BODY=phish", ""},       // mixed case second param
		{"mailto:a@b.com?subject=hello", "mailto:a@b.com?subject=hello"}, // subject still allowed
		{"//evil.com", ""}, // protocol-relative rejected
		{"javascript:alert(1)", ""},       // bad scheme
		{"/r/foo", "/r/foo"},              // relative
		{"./local", "./local"},
		{"#section", "#section"}, // fragment-only
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
		{"mailto:a@b.com", ""}, // mailto rejected for images
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
		{addrG, "", addrG},  // any prefix
		{addrG, "gpub", ""}, // wrong prefix
		{"gpub1abc123def456ghijklmn", "gpub", "gpub1abc123def456ghijklmn"},
		{"gpub1abc123def456ghijklmn", "", "gpub1abc123def456ghijklmn"},
		{"b1xyz789abc123def456", "", "b1xyz789abc123def456"}, // any-prefix mode allows b1...
		{"g1ABC", "g", ""}, // uppercase rejected
		{"x", "g", ""},
		{"", "g", ""},
	}
	for _, c := range cases {
		if got := BechString(c.s, c.prefix); got != c.want {
			t.Errorf("BechString(%q,%q) = %q, want %q", c.s, c.prefix, got, c.want)
		}
	}
}

func TestFootnoteLabel(t *testing.T) {
	cases := []struct{ in, want string }{
		{"note1", "note1"},
		{"Note_A-1", "Note_A-1"},
		{"with space", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := FootnoteLabel(c.in); got != c.want {
			t.Errorf("FootnoteLabel(%q) = %q, want %q", c.in, got, c.want)
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
