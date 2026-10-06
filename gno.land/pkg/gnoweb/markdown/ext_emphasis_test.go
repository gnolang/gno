package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
)

// newEmphasisGuardedMarkdown builds a goldmark instance with ONLY the default
// parsers plus the bounded emphasis guard — the minimal setup to exercise the
// guard in isolation.
func newEmphasisGuardedMarkdown() goldmark.Markdown {
	m := goldmark.New()
	ExtEmphasis.Extend(m)
	return m
}

func renderMarkdownString(t *testing.T, m goldmark.Markdown, src string) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, m.Convert([]byte(src), &buf))
	return buf.String()
}

// assertEmphasisGuardActive renders far more than the per-block cap of
// well-matched emphasis spans through m and asserts the guard clipped them: the
// rendered <em> count is bounded by the cap, not the input size. This is a
// deterministic proxy for "the bounded emphasis parser is active on this
// instance" — without it, goldmark's default parser would emphasize every span
// (and re-open the yuin/goldmark#555 quadratic). Deterministic, so it never
// flakes under -race or a loaded runner the way a wall-clock bound would.
func assertEmphasisGuardActive(t *testing.T, m goldmark.Markdown) {
	t.Helper()
	src := strings.Repeat("*x* ", MaxEmphasisDelimitersPerBlock*3)
	got := renderMarkdownString(t, m, src)
	n := strings.Count(got, "<em>")
	require.LessOrEqualf(t, n, MaxEmphasisDelimitersPerBlock, "<em> count = %d (emphasis guard not active on this instance)", n)
}

// emphasisGoldenCap keeps the golden fixtures tiny: with 4 delimiter runs per
// block, two spans sit exactly at the cap and the clip boundary is visible in
// a one-line input instead of a 1000-span blob.
const emphasisGoldenCap = 4

func testEmphasisGuardOutput(t *testing.T, nameIn string, input []byte) (string, []byte) {
	t.Helper()
	assertExt(t, nameIn, ".md")

	m := goldmark.New()
	NewExtEmphasis(emphasisGoldenCap).Extend(m)

	var html bytes.Buffer
	require.NoError(t, m.Convert(input, &html))
	return "output.html", html.Bytes()
}

// Golden files pin the exact HTML shape of the guard's behavior — where
// emphasis stops and literal '*' text begins — on a small-cap instance.
// The production-cap runner (TestGnoExtension) skips this directory.
func TestEmphasisGuardGolden(t *testing.T) {
	gold := NewGoldentTests(testEmphasisGuardOutput)
	gold.Update = *update
	gold.Run(t, "golden/ext_emphasis")
}

func TestBoundedEmphasis_NormalEmphasisUnchanged(t *testing.T) {
	m := newEmphasisGuardedMarkdown()
	cases := []struct{ in, want string }{
		{"*a*", "<em>a</em>"},
		{"**b**", "<strong>b</strong>"},
		{"_c_", "<em>c</em>"},
		{"__d__", "<strong>d</strong>"},
		{"a*b*c _d_ e", "<em>b</em>"},
	}
	for _, tc := range cases {
		assert.Contains(t, renderMarkdownString(t, m, tc.in), tc.want, "render(%q)", tc.in)
	}
}

// The 1 MB goal: a large multi-paragraph doc whose TOTAL emphasis-run count far
// exceeds the per-block cap, but no single block does. Every span must render —
// proving the per-block CloseBlock reset never clips legit content.
func TestBoundedEmphasis_LegitMultiBlockNotClipped(t *testing.T) {
	const paragraphs = MaxEmphasisDelimitersPerBlock // 1000 blocks, 4 runs each
	var sb strings.Builder
	for range paragraphs {
		sb.WriteString("This is *one* and **two** here.\n\n")
	}
	m := newEmphasisGuardedMarkdown()
	got := renderMarkdownString(t, m, sb.String())
	assert.Equal(t, paragraphs, strings.Count(got, "<em>"), "emphasis clipped across blocks")
	assert.Equal(t, paragraphs, strings.Count(got, "<strong>"))
}

// One block with far more than the cap of well-matched emphasis spans. Beyond
// the cap, markers render as literal text, so the <em> count is bounded by the
// cap (not proportional to input).
func TestBoundedEmphasis_OverCapSingleBlockClipped(t *testing.T) {
	assertEmphasisGuardActive(t, newEmphasisGuardedMarkdown())
}
