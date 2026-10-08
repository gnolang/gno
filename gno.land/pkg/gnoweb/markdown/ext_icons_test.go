package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// iconCatalogFile lists every icon and iconCatalogImage shows them: GitHub
// strips <gno-icon> from markdown, so the catalog embeds a plain SVG image
// generated from the same registry. TestIconCatalog keeps both in sync
// (`make icons` regenerates them).
const (
	iconCatalogFile  = "ICONS.md"
	iconCatalogImage = "ICONS.svg"
)

func TestParseIconTag(t *testing.T) {
	for _, tc := range []struct {
		line        string
		size        int
		name, label string
		open        bool // not self-closing
	}{
		{line: `<gno-icon name="star" /> tail`, size: 24, name: "star"},
		{line: `<gno-icon name="star"/>`, size: 23, name: "star"},
		{line: `<gno-icon name=" star " label=" a > b " />`, size: 42, name: "star", label: "a > b"},
		{line: `<GNO-ICON NAME="star" LABEL='x' />`, size: 34, name: "star", label: "x"},
		{line: `<gno-icon name=star />`, size: 22, name: "star"},
		{line: `<gno-icon name="a" name="b" label />`, size: 36, name: "a"},
		{line: `<gno-icon name="st&#97;r" />`, size: 28, name: "st&#97;r"},
		{line: `<gno-icon />`, size: 12},
		{line: `<gno-icon name="star">`, size: 22, name: "star", open: true},
		{line: `<gno-icon>`, size: 10, open: true},
		{line: `<gno-icon name="star`}, // unterminated; the shared scanner's own cases are in TestScanGnoTag
		{line: `<gno-icon label="` + strings.Repeat("x", maxIconTagLen) + `" />`},
	} {
		t.Run(tc.line, func(t *testing.T) {
			size, icon := parseIconTag([]byte(tc.line))
			assert.Equal(t, tc.size, size)
			if tc.size == 0 {
				return
			}
			assert.Equal(t, tc.name, string(icon.Name))
			assert.Equal(t, tc.label, string(icon.Label))
			assert.Equal(t, !tc.open, icon.SelfClosing)
		})
	}
}

func TestParseIconTagAllocs(t *testing.T) {
	line := []byte(`<gno-icon name="check-circle" label="Verified &amp; signed" /> tail`)
	assert.Zero(t, testing.AllocsPerRun(100, func() { parseIconTag(line) }))
}

// TestIconParseLinear renders inputs that made parsing quadratic when each
// unterminated `<gno-icon` was scanned to the end of its line: 200 KB in one
// line took 15 s. Bounded, each takes well under the budget.
func TestIconParseLinear(t *testing.T) {
	m := newProductionLikeMarkdown()
	for name, src := range map[string]string{
		"bare":       strings.Repeat("<gno-icon ", 20000),
		"open-quote": strings.Repeat(`<gno-icon name="`, 10000),
		"heading":    "## " + strings.Repeat("<gno-icon x ", 20000),
		"valid":      strings.Repeat(`<gno-icon name="star" />`, 10000),
	} {
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			var buf bytes.Buffer
			require.NoError(t, m.Convert([]byte(src), &buf, parser.WithContext(NewGnoParserContext(GnoContext{}))))
			assert.Less(t, time.Since(start), 2*time.Second)
		})
	}
}

// newProductionLikeMarkdown mirrors gnoweb's NewDefaultGoldmarkOptions (auto
// heading IDs and the GFM extensions), which the golden runner leaves out.
func newProductionLikeMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Table,
			NewGnoExtension(),
		),
	)
}

func TestIconHeadingIDAndToc(t *testing.T) {
	src := []byte(`# Title

## <gno-icon name="rocket" /> Launch

## Launch

## Status <gno-icon name="check-circle" label="done" />

## <gno-icon name="star" />

## <gno-icon name="star" label="Top" />

## ` + "`<gno-icon name=\"star\" />`" + ` syntax

## \<gno-icon name="star" /> escaped

## Plain heading
`)
	m := newProductionLikeMarkdown()
	doc := m.Parser().Parse(text.NewReader(src), parser.WithContext(NewGnoParserContext(GnoContext{})))
	var buf bytes.Buffer
	require.NoError(t, m.Renderer().Render(&buf, src, doc))
	out := buf.String()

	// IDs come from the visible text, deduplicated in document order.
	assert.Contains(t, out, `<h2 id="launch"><svg class="gno-icon"`)
	assert.Contains(t, out, `<h2 id="launch-1">Launch</h2>`)
	assert.Contains(t, out, `<h1 id="title">Title</h1>`)
	assert.Contains(t, out, `<h2 id="status">Status <svg class="gno-icon"`)
	assert.Contains(t, out, `<h2 id="heading"><svg class="gno-icon"`)
	assert.Contains(t, out, `<h2 id="heading-1"><svg class="gno-icon" viewBox="0 0 21 21" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round" role="img" aria-label="Top">`)
	// A tag shown as text (code span, backslash escape) stays in the ID.
	assert.Contains(t, out, `<h2 id="gno-icon-namestar--syntax"><code>`)
	assert.Contains(t, out, `<h2 id="gno-icon-namestar--escaped">&lt;gno-icon`)
	assert.Contains(t, out, `<h2 id="plain-heading">Plain heading</h2>`)
	assert.NotContains(t, out, "gno-icon-namerocket")

	toc, err := TocInspect(doc, src, TocOptions{MinDepth: 2, MaxDepth: 6})
	require.NoError(t, err)
	var got []string
	for _, item := range toc.Items {
		got = append(got, item.ID+"="+strings.TrimSpace(item.Title))
	}
	assert.Equal(t, []string{
		"launch=Launch",
		"launch-1=Launch",
		"status=Status",
		`gno-icon-namestar--syntax=<gno-icon name="star" /> syntax`,
		`gno-icon-namestar--escaped=<gno-icon name="star" /> escaped`,
		"plain-heading=Plain heading",
	}, got, "an icon-only heading has no text, so the TOC drops it, label or not")
}

func TestIconInGFM(t *testing.T) {
	src := []byte(`| Status | Name |
|---|---|
| <gno-icon name="check-circle" label="ok" /> | alpha |
| <gno-icon name="unicorn" /> | beta |
| <gno-icon name="star" label="a\|b" /> | gamma |
| <gno-icon name="star" label="c&#124;d" /> | delta |

~~<gno-icon name="trash" /> struck~~
`)
	var buf bytes.Buffer
	require.NoError(t, newProductionLikeMarkdown().Convert(src, &buf))
	out := buf.String()

	assert.Contains(t, out, `<td><svg class="gno-icon" viewBox="0 0 21 21" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round" role="img" aria-label="ok">`)
	assert.Contains(t, out, `<td><!-- gno-icon: unknown name "unicorn" --></td>`)
	assert.Contains(t, out, `<td><svg class="gno-icon" viewBox="0 0 21 21" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round" role="img" aria-label="a|b">`,
		"a | in a label is written \\| (or &#124;) inside a table cell")
	assert.Contains(t, out, `aria-label="c|d"`)
	assert.Contains(t, out, `<del><svg class="gno-icon"`)
}

// TestIconCatalog keeps ICONS.md and ICONS.svg in sync with the registry.
// Regenerate them with `make icons` (or -update-golden-tests).
func TestIconCatalog(t *testing.T) {
	names := slices.Sorted(maps.Keys(iconRegistry))

	var md bytes.Buffer
	md.WriteString(`# gnoweb icons

<!-- Generated by ` + "`make icons`" + ` from the icon registry. DO NOT EDIT. -->

Write ` + "`<gno-icon name=\"star\" />`" + ` anywhere inline markdown goes: a heading,
a paragraph, a list item, a table cell, a link label. Add
` + "`label=\"…\"`" + ` when the icon carries meaning on its own, so screen readers
announce it; leave it out for a decorative icon.

` + "```markdown" + `
## <gno-icon name="rocket" /> Launch

Status: <gno-icon name="check-circle" label="Verified" />

[<gno-icon name="book" /> Read the docs](/r/docs)
` + "```" + `

Rules:

- One line, self-closing: ` + "`<gno-icon name=\"star\" />`" + `. A tag that is not
  self-closing, or whose name is missing or unknown, renders nothing but an
  HTML comment saying why.
- Names are lowercase, exactly as listed below.
- Keep the tag under 512 bytes. In a table cell, write a ` + "`|`" + ` in a label as
  ` + "`\\|`" + `.
- A link or heading holding only an icon needs ` + "`label`" + `: it is its only name.
  Such a heading has no text, so it gets no table-of-contents entry.

To add an icon, list it in ` + "`icons/icons.txt`" + ` and run ` + "`make icons`" + `.

![Every icon with its name](ICONS.svg)

`)
	fmt.Fprintf(&md, "%d icons:\n\n| Icon | Name |\n|---|---|\n", len(names))
	for _, name := range names {
		fmt.Fprintf(&md, "| <gno-icon name=\"%s\" /> | `%s` |\n", name, name)
	}

	// A grid of every glyph over its name, dark on light so it reads in
	// either GitHub theme.
	const cols, cellW, cellH = 8, 120, 64
	rows := (len(names) + cols - 1) / cols
	var svg bytes.Buffer
	fmt.Fprintf(&svg, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" font-family="sans-serif" font-size="11" text-anchor="middle">`+"\n",
		cols*cellW, rows*cellH)
	svg.WriteString(`<rect width="100%" height="100%" fill="#fff"/>` + "\n")
	for i, name := range names {
		x, y := i%cols*cellW, i/cols*cellH
		glyph := iconRegistry[name]
		fmt.Fprintf(&svg, `<svg x="%d" y="%d" width="24" height="24" color="#1a1a1a" %s>%s</svg>`+"\n",
			x+cellW/2-12, y+10, glyph.head, glyph.body)
		fmt.Fprintf(&svg, `<text x="%d" y="%d" fill="#444">%s</text>`+"\n", x+cellW/2, y+52, name)
	}
	svg.WriteString("</svg>\n")

	checkGenerated(t, iconCatalogFile, md.Bytes())
	checkGenerated(t, iconCatalogImage, svg.Bytes())
}

// BenchmarkIconRender measures rendering one parsed icon: the registry
// lookup and the writes, nothing else.
func BenchmarkIconRender(b *testing.B) {
	for _, src := range []string{
		`<gno-icon name="star" />`,
		`<gno-icon name="check-circle" label="Verified" />`,
		`<gno-icon name="search" />`, // a chrome icon
	} {
		source := []byte(src)
		m := newProductionLikeMarkdown()
		doc := m.Parser().Parse(text.NewReader(source))
		icon := doc.FirstChild().FirstChild()
		require.Equal(b, KindIcon, icon.Kind())
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)
		b.Run(src, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				buf.Reset()
				renderIcon(w, source, icon, true)
				w.Flush()
			}
			b.ReportMetric(float64(buf.Len()), "bytes/icon")
		})
	}
}

func BenchmarkIconParse(b *testing.B) {
	m := newProductionLikeMarkdown()
	for _, bc := range []struct{ name, line string }{
		{"icons", `Status <gno-icon name="check-circle" label="ok" /> and <gno-icon name="star" />. `},
		// The common case: '<' that is not an icon must stay on the cheap
		// prefix check.
		{"other-tags", `Some <b>bold</b> and <i>raw</i> html <br/> here. `},
		// Unterminated tags: each scan stops at maxIconTagLen.
		{"unterminated", `<gno-icon name="star `},
	} {
		src := []byte(strings.Repeat(bc.line, 200))
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.Parser().Parse(text.NewReader(src))
			}
		})
	}
}
