package markdown

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// parseDetails parses src with the gno extensions and returns its HTML and
// the gno-* nesting depth left on the parser context.
func parseDetails(t *testing.T, src string) (string, int) {
	t.Helper()
	m := newGnoMarkdown()
	pc := NewGnoParserContext(GnoContext{})
	doc := m.Parser().Parse(text.NewReader([]byte(src)), parser.WithContext(pc))
	var buf bytes.Buffer
	require.NoError(t, m.Renderer().Render(&buf, []byte(src), doc))
	return buf.String(), Get(pc)
}

// A run of openers stops at the shared nesting cap; the rest are refused.
func TestDetailsNestBounded(t *testing.T) {
	out, _ := parseDetails(t, strings.Repeat("<gno-details summary=\"a\">\n", 4000))
	require.Equal(t, MaxGnoNestDepth, strings.Count(out, `<details class="gno-details">`))
}

// Every Push in Open is matched by one Pop, whatever shape ends the block.
func TestDetailsDepthBalanced(t *testing.T) {
	cases := map[string]string{
		"closed":          "<gno-details>\nb\n</gno-details>\n",
		"unclosed":        "<gno-details>\nb\n",
		"over_cap":        strings.Repeat("<gno-details>\n", 10),
		"in_quote_ends":   "> <gno-details>\n> b\n\nafter\n",
		"in_list_ends":    "- <gno-details>\n  b\n\npara\n",
		"inner_unclosed":  "<gno-details>\n<gno-details>\nb\n</gno-details>\n",
		"close_elsewhere": "<gno-details>\n- item\n  </gno-details>\n",
		"sequential":      strings.Repeat("<gno-details>\nb\n</gno-details>\n", 6),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			_, depth := parseDetails(t, src)
			require.Zero(t, depth)
		})
	}
}

// Closed blocks give their depth back while parsing, so a gno-* block
// after four of them still opens.
func TestDetailsSequentialKeepsCap(t *testing.T) {
	src := strings.Repeat("<gno-details>\nb\n</gno-details>\n", 5) + "> [!NOTE]\n> x\n"
	out, _ := parseDetails(t, src)
	require.Contains(t, out, "gno-alert")
}

// A block left unclosed in a container ends with it and gives its depth
// back, so the gno-* blocks after it still open.
func TestDetailsReleasedWithContainer(t *testing.T) {
	tail := "\n\n<gno-columns>\nA\n<gno-columns-sep>\nB\n</gno-columns>\n\n> [!WARNING]\n> x\n\n<gno-details summary=\"After\">\nb\n</gno-details>\n"
	cases := map[string]string{
		"quote": strings.Repeat("> <gno-details>\n", 4),
		"list":  strings.Repeat("- <gno-details>\n  b\n", 4),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			out, depth := parseDetails(t, src+tail)
			require.Contains(t, out, `class="gno-columns"`)
			require.Contains(t, out, "gno-alert")
			require.Contains(t, out, "<span>After</span>")
			require.Zero(t, depth)
		})
	}
}

// An opener straight after a paragraph in an ended list item releases the
// blocks left in it, so it opens rather than being refused at the cap.
func TestDetailsReleasedBeforeOpener(t *testing.T) {
	src := "- <gno-details>\n" + strings.Repeat("  <gno-details>\n", 3) + "  p\n" +
		"<gno-details summary=\"After\">\nb\n</gno-details>\n"
	out, depth := parseDetails(t, src)
	require.Contains(t, out, "<span>After</span>")
	require.Zero(t, depth)
}

// bestParseTime returns the fastest of three parses of src.
func bestParseTime(t *testing.T, src string) time.Duration {
	t.Helper()
	d := time.Duration(1<<63 - 1)
	for range 3 {
		runtime.GC()
		start := time.Now()
		parseDetails(t, src)
		d = min(d, time.Since(start))
	}
	return d
}

// A line of `>` markers below an unclosed block costs about what it costs
// without one: the release walks the quote chain once per line, not once
// per marker.
func TestDetailsQuoteMarkersLinear(t *testing.T) {
	line := strings.Repeat(">", 20000) + "\n"
	ratio := float64(bestParseTime(t, "<gno-details>\n"+line)) / float64(bestParseTime(t, line))
	assert.Less(t, ratio, 3.0, "the opener took %.1fx the time", ratio)
}

// Trailing whitespace after a line of `>` markers costs about what a
// trailing text byte costs: no marker rescans the rest of the line.
func TestDetailsQuoteMarkersTrailingSpaceLinear(t *testing.T) {
	line := strings.Repeat(">", 2000) + strings.Repeat(" ", 100000)
	ratio := float64(bestParseTime(t, line+"\n")) / float64(bestParseTime(t, line+"x\n"))
	assert.Less(t, ratio, 3.0, "the trailing spaces took %.1fx the time", ratio)
}

// A `<` line continuing a paragraph lazily keeps its container open, so the
// block in it pairs with its close tag.
func TestDetailsLazyLineKeepsBlock(t *testing.T) {
	out, _ := parseDetails(t, "> <gno-details>\n> p\n<i>x</i>\n> </gno-details>\n> after\n")
	require.Less(t, strings.Index(out, "</details>"), strings.Index(out, "<p>after</p>"))
}

// CRLF line endings parse as LF ones.
func TestDetailsCRLF(t *testing.T) {
	src := "<gno-details summary=\"S\" open>\nbody\n</gno-details>\nafter\n"
	lf, _ := parseDetails(t, src)
	crlf, _ := parseDetails(t, strings.ReplaceAll(src, "\n", "\r\n"))
	require.Equal(t, lf, crlf)
}

// An alert straight after stale blocks reads the depth before any `<` line
// comes, so the release has to run on its `>` line too.
func TestDetailsReleasedBeforeAlert(t *testing.T) {
	cases := map[string]string{
		"after_blank":     strings.Repeat("> <gno-details>\n", 4) + "\n> [!WARNING]\n> x\n",
		"after_paragraph": "- <gno-details>\n" + strings.Repeat("  <gno-details>\n", 3) + "  p\n> [!WARNING]\n> x\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			out, _ := parseDetails(t, src)
			require.Contains(t, out, "gno-alert")
		})
	}
}

// A block left unclosed in a column ends at the next grid tag, where the
// transformer stops it, so four alerts nested after the grid all open.
func TestDetailsEndsAtGrid(t *testing.T) {
	src := "<gno-columns>\n<gno-details>\nleft\n<gno-columns-sep>\nB\n</gno-columns>\n\n" +
		"> [!NOTE]\n> > [!NOTE]\n> > > [!NOTE]\n> > > > [!NOTE]\n> > > > x\n"
	out, depth := parseDetails(t, src)
	require.Equal(t, 4, strings.Count(out, `class="gno-alert `))
	require.Zero(t, depth)
}
