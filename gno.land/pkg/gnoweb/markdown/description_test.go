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
		{"a document with no paragraph has no summary", "# Only\n\n## Headings\n", ""},
		{"empty", "", ""},
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
			assert.Equal(t, tc.want, Description(doc, src))
		})
	}
}

func TestDescriptionTruncates(t *testing.T) {
	t.Parallel()

	src := []byte(strings.Repeat("word ", 60))
	doc := goldmark.New().Parser().Parse(text.NewReader(src))
	got := Description(doc, src)

	assert.LessOrEqual(t, len([]rune(got)), descriptionMaxRunes+1, "one rune of headroom for the ellipsis")
	assert.True(t, strings.HasSuffix(got, "…"), "a cut summary must say it was cut")
	assert.False(t, strings.HasSuffix(strings.TrimSuffix(got, "…"), " "), "the cut must not leave a trailing space")
}
