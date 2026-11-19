package markdown

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSummary(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("word ", 60) // 300 runes, no sentence end

	cases := []struct {
		name, src string
		limit     int
		want      string
	}{
		{"empty", "", 240, ""},
		{"one sentence", "Small on-chain ASCII art generator.", 240, "Small on-chain ASCII art generator."},
		{
			"paragraphs join",
			"Boards2 is a forum.\n\nUsers start threads.",
			240,
			"Boards2 is a forum. Users start threads.",
		},
		{
			"stops at a list",
			"Pools for GnoSwap.\n\n- Concentrated liquidity\n- Fees",
			240,
			"Pools for GnoSwap.",
		},
		{
			"keeps whole sentences only",
			"First sentence here. Second sentence is longer than what fits.",
			30,
			"First sentence here.",
		},
		{"cuts a lone long sentence on a word", long, 20, "word word word word…"},
		{"keeps link text, drops target", "A demo of [b58](/p/moul/b58/v0) codecs.", 240, "A demo of b58 codecs."},
		{"drops image alt text", "![secret](x.png) Visible text.", 240, "Visible text."},
		{"drops bidi overrides", "Safe‮ text.", 240, "Safe text."},
		{"stops at a heading", "Intro.\n\n# Usage\n\nCall it.", 240, "Intro."},
		{"parses only the top of a long document", "Top." + strings.Repeat(" x", summaryMaxBytes), 10, "Top."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, Summary(tc.src, tc.limit))
		})
	}
}
