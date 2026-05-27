package markdown

import (
	"strings"
	"testing"
)

func TestStripBidiAndZeroWidth(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"a\u200Bb", "ab"},          // ZWSP
		{"a\u202Eevil\u202Cb", "aevilb"}, // RLO + PDF
		{"\uFEFFstart", "start"},    // BOM
		{"clean\nbreak", "clean\nbreak"},
	}
	for _, c := range cases {
		if got := StripBidiAndZeroWidth(c.in); got != c.want {
			t.Errorf("StripBidiAndZeroWidth(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeBreaks(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a\nb", "a\nb"},
		{"a\rb", "a\nb"},
		{"a\r\nb", "a\nb"},
		{"a\r\n\rb", "a\n\nb"},
	}
	for _, c := range cases {
		if got := NormalizeBreaks(c.in); got != c.want {
			t.Errorf("NormalizeBreaks(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEscapeInline(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"a*b", `a\*b`},
		{"a|b", "a|b"}, // | NOT in inline set
		{"a=b", "a=b"}, // = NOT in inline set
		{"a\x00b", "a\xef\xbf\xbdb"},
		{"\\*", `\\\*`},
	}
	for _, c := range cases {
		if got := EscapeInline(c.in); got != c.want {
			t.Errorf("EscapeInline(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEscapeTitle(t *testing.T) {
	cases := []struct{ in, want string }{
		{`he said "hi"`, `he said \"hi\"`},
		{`it's a (test)`, `it\'s a \(test\)`},
	}
	for _, c := range cases {
		if got := EscapeTitle(c.in); got != c.want {
			t.Errorf("EscapeTitle(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPercentEncodeURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"http://a.com/x", "http://a.com/x"},
		{"a b", "a%20b"},
		{"a<b", "a%3Cb"},
		{"a%20b", "a%20b"}, // already encoded, preserved
		{"a%zzb", "a%25zzb"}, // bare %
	}
	for _, c := range cases {
		if got := PercentEncodeURL(c.in); got != c.want {
			t.Errorf("PercentEncodeURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMatchCharsetN(t *testing.T) {
	// Build bitmaps for [a-z][a-z0-9]* with bounds [1, 64].
	var firstLo, firstHi uint64
	for c := byte('a'); c <= 'z'; c++ {
		if c < 64 {
			firstLo |= 1 << c
		} else {
			firstHi |= 1 << (c - 64)
		}
	}
	restLo, restHi := firstLo, firstHi
	for c := byte('0'); c <= '9'; c++ {
		restLo |= 1 << c
	}
	cases := []struct {
		s    string
		want bool
	}{
		{"abc", true},
		{"abc123", true},
		{"1abc", false}, // first must be letter
		{"", false},     // below minLen
		{"a_b", false},  // _ not in set
	}
	for _, c := range cases {
		if got := MatchCharsetN(c.s, firstLo, firstHi, restLo, restHi, 1, 64); got != c.want {
			t.Errorf("MatchCharsetN(%q) = %v, want %v", c.s, got, c.want)
		}
	}
}

func TestCodeFence(t *testing.T) {
	cases := []struct {
		content string
		min     int
		want    string
	}{
		{"", 3, "```"},
		{"```", 3, "````"},                  // 3 backticks → 4-fence
		{"a `b` c", 3, "```"},               // longest run is 1, min wins
		{"a ``` b ```` c", 3, "`````"},      // longest run 4, +1 = 5
		{"", 0, "`"},                        // min < 1 clamped to 1
	}
	for _, c := range cases {
		if got := CodeFence(c.content, c.min); got != c.want {
			t.Errorf("CodeFence(%q, %d) = %q, want %q", c.content, c.min, got, c.want)
		}
	}
}

func TestEscapeBlockHazards(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello world\n", "hello world\n"},
		{"atx-heading", "# evil heading\n", "\\# evil heading\n"},
		{"blockquote", "> attacker\n", "\\> attacker\n"},
		{"bullet-list", "- item\n", "\\- item\n"},
		{"ordered-list", "1. item\n", "1\\. item\n"},
		{"fence-open-autoclose", "```\nuser code\n", "```\nuser code\n```\n"},
		{"fence-roundtrip", "```\nuser code\n```\n", "```\nuser code\n```\n"},
		{"setext-h1", "title\n===\n", "title\n\\===\n"},
		{"ref-link-use", "[click][evil]\n", "\\[click\\]\\[evil\\]\n"},
		{"shortcut-ref", "[label]\n", "\\[label\\]\n"},
		{"footnote-ref", "[^name]\n", "\\[^name\\]\n"},
		{"inline-link-preserved", "[text](url)\n", "[text](url)\n"},
		{"inline-image-preserved", "![alt](src)\n", "![alt](src)\n"},
		{"multi-line-lrd-stripped", "[lab\nel]: https://e.com\n\nbody\n", "\nbody\n"},
		{"backslash-escaped-lrd-not-stripped", "[label\\]: url\n", "\\[label\\]: url\n"},
		{"lrd-strip", "[evil]: https://bad\n", ""},
		{"u2028-fold", "a\u2028b\n", "a\nb\n"},
		{"nel-fold", "a\u0085b\n", "a\nb\n"},
		{"ext-delimiter", "<gno-card>\n", "\\<gno-card>\n"},
		{"gfm-table-row", "| a | b |\n", "\\| a | b |\n"},
	}
	for _, c := range cases {
		if got := EscapeBlockHazards(c.in); got != c.want {
			t.Errorf("%s: EscapeBlockHazards(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

// BenchmarkEscapeBlockHazardsAdversarial exercises bracket-walker
// inputs that maximize backtrack work in pass 1: repeated well-formed
// links (linear path), repeated unclosed-link prefixes terminated by a
// blank line (current super-linear path; see note), and a multi-line
// LRD label that must scan across newlines. The intent is a wall-clock
// budget guard against future regressions.
//
// NOTE: `unclosed-xN` currently scales worse than linear because every
// `[` re-scans forward to the blank-line terminator. For N=500/1k/2k/4k
// the per-op cost grows ~3x per 2x input. Acceptable for realistic
// realm input sizes (a few KB) but flagged here for follow-up if a
// caller needs to defend against pathological inputs.
func BenchmarkEscapeBlockHazardsAdversarial(b *testing.B) {
	links := strings.Repeat("[a](b)", 1000)
	unclosed := strings.Repeat("[a", 1000) + "\n\nbody"
	mlLRD := "[" + strings.Repeat("lab\nel", 500) + "]: url\n"
	cases := []struct {
		name string
		in   string
	}{
		{"links-x1000", links},
		{"unclosed-x1000", unclosed},
		{"multiline-lrd-label", mlLRD},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(c.in)))
			for i := 0; i < b.N; i++ {
				_ = EscapeBlockHazards(c.in)
			}
		})
	}
}
