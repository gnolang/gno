package markdown

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

func TestDescription(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			"first paragraph after the title",
			"# Title\n\nWhat this realm does, in one sentence long enough to be a summary.\n\nMore.\n",
			"What this realm does, in one sentence long enough to be a summary.",
		},
		{
			"inline markup is flattened",
			"A [link](/r/x) and **bold**, written out at summary length for once.\n",
			"A link and bold, written out at summary length for once.",
		},
		{
			"line breaks collapse",
			"A sentence that runs\nacross two source lines and stays one summary.\n",
			"A sentence that runs across two source lines and stays one summary.",
		},
		{
			"a paragraph too short to summarise is skipped",
			"16 Sep 2026\n\nThe post itself, which says what happened and why it matters.\n",
			"The post itself, which says what happened and why it matters.",
		},
		{"an index of dates has no summary", "# Blog\n\n16 Sep 2026\n\n27 Aug 2026\n", ""},
		{
			"a paragraph before any heading summarises the page",
			"What this realm does, in one sentence long enough to be a summary.\n\n# Title\n",
			"What this realm does, in one sentence long enough to be a summary.",
		},
		{
			// Like r/gov/dao: the lead is a title, and the prose sits in sections.
			"a paragraph under a later heading is a section's, not the page's",
			"# GovDAO\n## Proposals\n\nAuthor: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 proposed this.\n",
			"",
		},
		{
			"a rule ends the lead",
			"# Title\n\n---\n\nA paragraph long enough to summarise, but under the rule.\n",
			"",
		},
		{
			// Past the lead, the text belongs to whoever the page quotes.
			"text further down never summarises the page",
			"Short.\n\n- item\n\n## Sub\n\n> quote\n\nUser post: Claim your free GNOT airdrop at evil.example right now!\n\n# User-chosen title evil.example\n",
			"",
		},
		{"a document with no paragraph has no summary", "# Only\n\n## Headings\n", ""},
		{"empty", "", ""},
		{
			// The summary reads as the page does: entities and escapes
			// resolved, a code span as typed.
			name: "entities and escapes are resolved, code spans are not",
			src:  "Tom &amp; Jerry cost 5 &euro; each, \\*not\\* 10, unlike `a &amp; b`.\n",
			want: "Tom & Jerry cost 5 € each, *not* 10, unlike a &amp; b.",
		},
		{
			name: "an escaped entity stays literal, as on the page",
			src:  "An escaped \\&amp; stays as typed on the page, and in the summary too.\n",
			want: "An escaped &amp; stays as typed on the page, and in the summary too.",
		},
		{
			// An alt shows only when the image fails, so lifting it would
			// publish a summary the reader never sees.
			name: "an image alt is not a summary",
			src:  "# H\n\n![" + strings.Repeat("hidden ", 12) + "](x.png)\n\nShort.\n",
			want: "",
		},
		{
			name: "text around an image survives without the alt",
			src:  "# H\n\nText before ![alt text](x.png) and text after, long enough to pass.\n",
			want: "Text before and text after, long enough to pass.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src := []byte(tc.src)
			doc := goldmark.New().Parser().Parse(text.NewReader(src))
			_, got := Lead(doc, src)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTitle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		src  string
		want string
	}{
		{"a leading h1", "# The post title\n\nIntro.\n\n# Another\n", "The post title"},
		// Further down, an h1 may head a comment or a quoted post.
		{"an h1 after other text is not the page's", "Intro.\n\n# The post title\n", ""},
		{"a banner image before the h1 does not count", "![banner](x.png)\n\n# Named\n", "Named"},
		{"an h2 is not a title", "## Section\n\nText.\n", ""},
		{"an empty leading h1 is no title", "#\n\n# Named\n", ""},
		{"inline markup is flattened", "# A [link](/r/x) and **bold**\n", "A link and bold"},
		{"entities and escapes are resolved", "# Tom &amp; Jerry \\*live\\*\n", "Tom & Jerry *live*"},
		{"an image alt is not a title", "# ![Official notice](x.png)\n", ""},
		// A bidi override reorders what a tab shows; a zero-width space hides.
		{"format characters are dropped", "# Pay \u202eelpmaxe.live\u202c to \u200bgno\n", "Pay elpmaxe.live to gno"},
		{"a nested h1 is not the page's", "> # Quoted\n", ""},
		// Backing off to the last space would leave "Welcome to…", which
		// names nothing; a long word is cut inside instead.
		{
			"a long unbroken word is cut, not dropped",
			"# Welcome to gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/home\n",
			"Welcome to gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsq…",
		},
		{"empty", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src := []byte(tc.src)
			doc := goldmark.New().Parser().Parse(text.NewReader(src))
			got, _ := Lead(doc, src)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("truncates", func(t *testing.T) {
		t.Parallel()

		src := []byte("# " + strings.Repeat("word ", 30))
		doc := goldmark.New().Parser().Parse(text.NewReader(src))
		got, _ := Lead(doc, src)
		assert.LessOrEqual(t, len([]rune(got)), titleMaxRunes+1, "one rune of headroom for the ellipsis")
		assert.True(t, strings.HasSuffix(got, "…"))
	})
}

func TestDescriptionTruncates(t *testing.T) {
	t.Parallel()

	src := []byte(strings.Repeat("word ", 60))
	doc := goldmark.New().Parser().Parse(text.NewReader(src))
	_, got := Lead(doc, src)

	assert.LessOrEqual(t, len([]rune(got)), descriptionMaxRunes+1, "one rune of headroom for the ellipsis")
	assert.True(t, strings.HasSuffix(got, "…"), "a cut summary must say it was cut")
	assert.False(t, strings.HasSuffix(strings.TrimSuffix(got, "…"), " "), "the cut must not leave a trailing space")
}

// A paragraph of many small text nodes resolves each one; a full-size write
// buffer per node made the lead cost several times the render itself.
func TestResolveTextAllocatesLittle(t *testing.T) {
	res := testing.Benchmark(func(b *testing.B) {
		b.Helper()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			resolveText([]byte("Tom &amp; Jerry"))
		}
	})
	assert.Less(t, res.AllocedBytesPerOp(), int64(256))
}

func BenchmarkLeadEmphasis(b *testing.B) {
	src := []byte(strings.Repeat("*word* ", 20000) + "\n")
	doc := goldmark.New().Parser().Parse(text.NewReader(src))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lead(doc, src)
	}
}
