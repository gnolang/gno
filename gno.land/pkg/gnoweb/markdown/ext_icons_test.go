package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"maps"
	"regexp"
	"runtime"
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
	fill := maxIconTagLen - len(`<gno-icon label="" />`) // label bytes for a 512-byte tag
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
		// The bound is inclusive: a 512-byte tag is read, a 513-byte one is not.
		{line: `<gno-icon label="` + strings.Repeat("x", fill) + `" />`, size: maxIconTagLen, label: strings.Repeat("x", fill)},
		{line: `<gno-icon label="` + strings.Repeat("x", fill+1) + `" />`},
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

// TestIconHeadingIDsLinear renders many identical headings holding icons.
// iconHeadingIDTransformer renumbers every heading once a heading holds an
// icon; with goldmark's generator that doubled its quadratic deduplication
// (20,000 took 35 s). linearIDs keeps it linear; plain duplicate headings
// are covered by TestDuplicateHeadingIDsLinear.
func TestIconHeadingIDsLinear(t *testing.T) {
	for name, gen := range map[string]func(n int) string{
		"icon-only":      func(n int) string { return strings.Repeat("## <gno-icon name=\"star\" />\n", n) },
		"icon-text":      func(n int) string { return strings.Repeat("## <gno-icon name=\"star\" /> a\n", n) },
		"after-one-icon": func(n int) string { return "## <gno-icon name=\"star\" /> b\n" + strings.Repeat("## a\n", n) },
	} {
		t.Run(name, func(t *testing.T) { assertLinear(t, gen, 5000) })
	}
}

// assertLinear renders gen(n) and gen(4n) and fails when the larger input
// takes more than 10 times as long: 4 times is linear, 16 quadratic. Comparing
// two sizes on the same machine, rather than against a fixed budget, holds
// under -race and on slow runners. Each size keeps its fastest of 5 runs,
// each after a GC, so a collection or a scheduler pause in one run does not
// skew the ratio (a CI runner hit 9x on a linear input with 3 runs).
func assertLinear(t *testing.T, gen func(n int) string, n int) {
	t.Helper()
	m := newProductionLikeMarkdown()
	measure := func(src []byte) time.Duration {
		best := time.Duration(1<<63 - 1)
		for range 5 {
			var buf bytes.Buffer
			runtime.GC()
			start := time.Now()
			require.NoError(t, m.Convert(src, &buf, parser.WithContext(NewGnoParserContext(GnoContext{}))))
			best = min(best, time.Since(start))
		}
		return best
	}
	small, large := measure([]byte(gen(n))), measure([]byte(gen(4*n)))
	// Below a few milliseconds the ratio is noise; nothing quadratic stays there.
	if large < 20*time.Millisecond {
		return
	}
	assert.Less(t, large, 10*small, "n=%d took %v, 4n took %v", n, small, large)
}

func TestRenderIconAllocs(t *testing.T) {
	m := newProductionLikeMarkdown()
	for _, src := range []string{
		`<gno-icon name="star" />`,
		`<gno-icon name="check-circle" label="Tom &amp; Jerry \| ok" />`,
	} {
		source := []byte(src)
		doc := m.Parser().Parse(text.NewReader(source))
		icon := doc.FirstChild().FirstChild()
		w := bufio.NewWriter(io.Discard)
		assert.Zero(t, testing.AllocsPerRun(100, func() { renderIcon(w, source, icon, true) }), src)
	}
}

// TestIconCRLF pins CRLF input, which goldens cannot hold (editors normalize
// line endings): the icon renders, the next line stays, a tag split across
// lines is not an icon.
func TestIconCRLF(t *testing.T) {
	src := "<gno-icon name=\"star\" />\r\nnext\r\n<gno-icon\r\nname=\"star\" />\r\n"
	var buf bytes.Buffer
	require.NoError(t, newProductionLikeMarkdown().Convert([]byte(src), &buf))
	out := buf.String()
	assert.Equal(t, 1, strings.Count(out, `<svg class="gno-icon"`), out)
	assert.Contains(t, out, "next")
	assert.NotContains(t, out, "\r<", "no stray CR inside the tag output")
}

// FuzzIconRender renders arbitrary input around icon tags and checks the
// invariants: no panic, a scanned tag is at most maxIconTagLen bytes, and
// every <svg> written is a gno-icon (or a chrome <use> glyph) carrying no
// event handler.
func FuzzIconRender(f *testing.F) {
	for _, seed := range []string{
		`<gno-icon name="star" />`,
		`<gno-icon name="check-circle" label="a > b" />`,
		`<gno-icon name="x" label='"><script>' onload=x />`,
		`## <gno-icon name="rocket" /> Launch`,
		`[<gno-icon name="globe" />](https://e.x)`,
		"<gno-icon name=\"star\r\n\" />",
		`<GNO-ICON NAME=star/>`,
		`<gno-icon name="star" label="turn on" />`,
		`<gno-icon name="star" label="x onclick=y" />`,
		`<gno-icon name="` + strings.Repeat("a", 600),
	} {
		f.Add(seed)
	}
	m := newProductionLikeMarkdown()
	f.Fuzz(func(t *testing.T, src string) {
		if size, _ := parseIconTag([]byte(src)); size > maxIconTagLen {
			t.Fatalf("tag of %d bytes read past maxIconTagLen", size)
		}
		var buf bytes.Buffer
		if err := m.Convert([]byte(src), &buf, parser.WithContext(NewGnoParserContext(GnoContext{}))); err != nil {
			return
		}
		for _, svg := range strings.Split(buf.String(), "<svg")[1:] {
			// gno-icons, or the chrome glyphs other extensions write
			// (alert kinds, link and form markers), which are <use> refs.
			if !strings.HasPrefix(svg, ` class="gno-icon" `) && !strings.Contains(svg[:min(len(svg), 40)], "<use href=\"#ico-") {
				t.Fatalf("an <svg> that is neither a gno-icon nor a chrome glyph: %q", svg[:min(len(svg), 80)])
			}
			// Attribute names only: a label value may hold " on".
			tag, _, _ := strings.Cut(svg, ">")
			toks, err := ParseHTMLTokens(strings.NewReader("<svg" + tag + ">"))
			require.NoError(t, err)
			for _, a := range toks[0].Attr {
				if strings.HasPrefix(a.Key, "on") {
					t.Fatalf("event handler in %q", tag)
				}
			}
		}
	})
}

func TestParseIconTagAllocs(t *testing.T) {
	line := []byte(`<gno-icon name="check-circle" label="Verified &amp; signed" /> tail`)
	assert.Zero(t, testing.AllocsPerRun(100, func() { parseIconTag(line) }))
}

// TestIconParseLinear renders inputs that made parsing quadratic when each
// unterminated `<gno-icon` was scanned to the end of its line: 200 KB in one
// line took 15 s. Bounded, each grows linearly with its size.
func TestIconParseLinear(t *testing.T) {
	const icon = `<gno-icon name="star" />`
	for name, gen := range map[string]func(n int) string{
		"bare":               func(n int) string { return strings.Repeat("<gno-icon ", 2*n) },
		"open-quote":         func(n int) string { return strings.Repeat(`<gno-icon name="`, n) },
		"heading":            func(n int) string { return "## " + strings.Repeat("<gno-icon x ", 2*n) },
		"lines-unterminated": func(n int) string { return strings.Repeat("<gno-icon name=\"x\n", 2*n) },
		"table-cells": func(n int) string {
			return "| a | b |\n|---|---|\n" + strings.Repeat("| <gno-icon name=\"x | y |\n", n)
		},
		"valid":            func(n int) string { return strings.Repeat(icon, n) },
		"heading-icons":    func(n int) string { return "## " + strings.Repeat(icon, n) },
		"link-icons":       func(n int) string { return "[" + strings.Repeat(icon, n) + "](/r/x)" },
		"link-filler-code": func(n int) string { return "[" + strings.Repeat("` ` ", 2*n) + strings.Repeat(icon, 2*n) + "](/r/x)" },
		"link-filler-html": func(n int) string {
			return "[" + strings.Repeat("<b></b>", 2*n) + strings.Repeat(icon, 2*n) + "](/r/x)"
		},
	} {
		t.Run(name, func(t *testing.T) {
			n := 2500
			if strings.Contains(gen(1), icon) {
				n = MaxIconsPerConvert / 8 // keep 4n under the cap, or a per-icon walk looks linear
			}
			assertLinear(t, gen, n)
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
	assert.Contains(t, out, `<h2 id="status-done">Status <svg class="gno-icon"`)
	assert.Contains(t, out, `<h2 id="heading"><svg class="gno-icon"`)
	assert.Contains(t, out, `<h2 id="top"><svg class="gno-icon" viewBox="0 0 21 21" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round" role="img" aria-label="Top">`)
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
		"status-done=Status done",
		"top=Top",
		`gno-icon-namestar--syntax=<gno-icon name="star" /> syntax`,
		`gno-icon-namestar--escaped=<gno-icon name="star" /> escaped`,
		"plain-heading=Plain heading",
	}, got, "a label names its heading in the TOC; an unlabeled icon-only heading has no title, so the TOC drops it")
}

// TestIconHeadingLabelText checks the TOC title and the ID of a heading
// take an icon's label only when the icon renders, decoded as its
// aria-label is, and apart from the text next to it.
func TestIconHeadingLabelText(t *testing.T) {
	m := newProductionLikeMarkdown()
	for src, want := range map[string]string{
		"## <gno-icon name=\"nope\" label=\"Top\" />\n":             "heading=",
		"## <gno-icon label=\"Top\" />\n":                           "heading=",
		"## <gno-icon name=\"star\" label=\"Top\">\n":               "heading=",
		"## <gno-icon name=\"star\" label=\"Q &amp; A\" />\n":       "q--a=Q & A",
		"## <gno-icon name=\"star\" label=\"&quot;Hi&quot;\" />\n":  "hi=\"Hi\"",
		"## Picks<gno-icon name=\"star\" label=\"Top\" />\n":        "picks-top=Picks Top",
		"## <gno-icon name=\"star\" label=\"Top\" />Picks\n":        "top-picks=Top Picks",
		"## A <gno-icon name=\"star\" label=\"Top\" /> B\n":         "a-top-b=A Top B",
		"## ![<gno-icon name=\"star\" label=\"Top\" />](i.png) A\n": "ipng-a= A", // the image source stays in the ID, as goldmark leaves it
	} {
		doc := m.Parser().Parse(text.NewReader([]byte(src)), parser.WithContext(NewGnoParserContext(GnoContext{})))
		var buf bytes.Buffer
		require.NoError(t, m.Renderer().Render(&buf, []byte(src), doc))
		toc, err := TocInspect(doc, []byte(src), TocOptions{MinDepth: 2, MaxDepth: 6})
		require.NoError(t, err)
		id := regexp.MustCompile(`<h2 id="([^"]*)"`).FindStringSubmatch(buf.String())
		require.Len(t, id, 2, "%q: %s", src, buf.String())
		title := ""
		if len(toc.Items) > 0 {
			title = toc.Items[0].Title
		}
		assert.Equal(t, want, id[1]+"="+title, "%q", src)
	}
}

// TestIconHeadingIDEmpty checks an empty heading keeps its ID in the
// sequence the icon transformer rebuilds, as goldmark numbers it.
func TestIconHeadingIDEmpty(t *testing.T) {
	m := newProductionLikeMarkdown()
	headingID := regexp.MustCompile(`<h2 id="([^"]*)"`)
	for src, want := range map[string][]string{
		"##\n\n## Heading\n": {"heading", "heading-1"},
		"##\n\n## Heading\n\n## <gno-icon name=\"star\" /> A\n":                  {"heading", "heading-1", "a"},
		"##\n\n## <gno-icon name=\"star\" label=\"Top\" />\n":                    {"heading", "top"},
		"## <gno-icon name=\"star\" />\n\n##\n\n## <gno-icon name=\"star\" />\n": {"heading", "heading-1", "heading-2"},
	} {
		var buf bytes.Buffer
		require.NoError(t, m.Convert([]byte(src), &buf, parser.WithContext(NewGnoParserContext(GnoContext{}))))
		var got []string
		for _, id := range headingID.FindAllStringSubmatch(buf.String(), -1) {
			got = append(got, id[1])
		}
		assert.Equal(t, want, got, "%q", src)
	}
}

// TestIconHintPerParent checks each link or heading gets its own hint: an
// icon in a heading does not answer for an icon-only link inside it.
func TestIconHintPerParent(t *testing.T) {
	const hint = `<!-- gno-icon: alone in a link or heading`
	m := newProductionLikeMarkdown()
	for src, want := range map[string]int{
		"## Title [<gno-icon name=\"star\" />](/r/x)\n":                                    1,
		"## <gno-icon name=\"star\" /> Title [<gno-icon name=\"star\" />](/r/x)\n":         1,
		"## <gno-icon name=\"star\" /> [<gno-icon name=\"star\" />](/r/x)\n":               2,
		"## [<gno-icon name=\"star\" />](/r/x) <gno-icon name=\"star\" />\n":               2,
		"## <gno-icon name=\"star\" /> [Title](/r/x)\n":                                    0,
		"## <gno-icon name=\"star\" /> [<gno-icon name=\"star\" label=\"Top\" />](/r/x)\n": 0,
		// An icon in an image's alt text renders nothing: it takes no hint.
		"## ![<gno-icon name=\"star\" />](i.png) <gno-icon name=\"star\" />\n": 1,
		"## ![<gno-icon name=\"star\" />](i.png)\n":                            0,
	} {
		var buf bytes.Buffer
		require.NoError(t, m.Convert([]byte(src), &buf))
		assert.Equal(t, want, strings.Count(buf.String(), hint), "%q:\n%s", src, buf.String())
	}
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

// TestIconBudget pins the icon cap: one Convert renders at most
// MaxIconsPerConvert icons, <gno-foreign> bodies included, and the tags past
// it take goldmark's raw HTML path, as without the extension. Uncapped, 1 MiB
// of `<gno-icon name=pure />` made a 105 MB page.
func TestIconBudget(t *testing.T) {
	const tag = "<gno-icon name=pure /> "
	render := func(src string) string {
		var buf bytes.Buffer
		require.NoError(t, newProductionLikeMarkdown().Convert([]byte(src), &buf))
		return buf.String()
	}

	flood := strings.Repeat(tag, (1<<20)/len(tag))
	out := render(flood)
	assert.Equal(t, MaxIconsPerConvert, strings.Count(out, "<svg"))
	assert.Less(t, len(out), 4<<20)

	// The budget is shared with a foreign body: half outside, the rest inside.
	half := MaxIconsPerConvert / 2
	out = render(strings.Repeat(tag, half) + "\n\n<gno-foreign>\n" + strings.Repeat(tag, MaxIconsPerConvert) + "\n</gno-foreign>\n")
	assert.Equal(t, MaxIconsPerConvert, strings.Count(out, "<svg"))
}

// TestIconLinkBadgesNotCallable pins the link badges out of the table:
// getLinkIcons drops them from <gno-foreign> links, so foreign content must
// not be able to add them back as icons.
func TestIconLinkBadgesNotCallable(t *testing.T) {
	for _, name := range []string{"external-link", "internal-link", "tx-link", "user-link"} {
		var buf bytes.Buffer
		src := "<gno-foreign>\n[Sign <gno-icon name=\"" + name + "\" />](/r/x$help&func=Foo)\n</gno-foreign>\n"
		require.NoError(t, newProductionLikeMarkdown().Convert([]byte(src), &buf))
		assert.NotContains(t, buf.String(), "<svg", name)
		assert.Contains(t, buf.String(), `<!-- gno-icon: unknown name "`+name+`" -->`)
	}
}
