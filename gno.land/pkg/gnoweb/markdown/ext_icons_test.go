package markdown

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
)

// iconCatalogFile lists every icon; TestIconCatalog keeps it in sync with the
// registry (`make icons` regenerates it).
const iconCatalogFile = "ICONS.md"

func TestIconRegistry(t *testing.T) {
	require.GreaterOrEqual(t, len(iconRegistry), 200)

	for name, glyph := range iconRegistry {
		require.NotEmpty(t, name)
		assert.True(t, bytes.HasPrefix(glyph.open, []byte(`<svg class="gno-icon" viewBox="`)), "%s: %s", name, glyph.open)
		assert.True(t, bytes.HasSuffix(glyph.body, []byte("</svg>")), name)

		// The glyph is re-serialized through the allowlist, so whatever the
		// source files hold, the body is shape elements with shape attributes.
		toks, err := ParseHTMLTokens(bytes.NewReader(glyph.body))
		require.NoError(t, err, name)
		for _, tok := range toks {
			switch tok.Type {
			case html.TextToken:
				assert.Empty(t, strings.TrimSpace(tok.Data), "%s: text in glyph", name)
			case html.EndTagToken:
				assert.True(t, tok.Data == "svg" || iconElements[tok.Data], "%s: </%s>", name, tok.Data)
			default:
				assert.True(t, iconElements[tok.Data], "%s: <%s>", name, tok.Data)
				for _, a := range tok.Attr {
					assert.True(t, iconAttrs[a.Key], "%s: <%s %s>", name, tok.Data, a.Key)
					assert.NotContains(t, a.Val, "url(", name)
				}
			}
		}
	}

	for name := range iconExcluded {
		assert.NotContains(t, iconRegistry, name)
	}
}

// TestIconRegistryNoShadowing checks that no source file reuses a name an
// earlier one defines, which would silently keep the earlier glyph.
func TestIconRegistryNoShadowing(t *testing.T) {
	seen := map[string]string{}
	for _, src := range iconSources {
		reg := map[string]iconGlyph{}
		require.NoError(t, loadIconSource(reg, src))

		for name := range reg {
			if prev, ok := seen[name]; ok {
				t.Errorf("%q in %s is shadowed by %s", name, src.name, prev)
			}
			seen[name] = src.name
		}
	}
}

func TestLoadIconsAllowlist(t *testing.T) {
	const src = `<svg>
<symbol id="ico-ok" viewBox="0 0 4 4" fill="none" onload="alert(1)" style="x">
  <title>Ok</title>
  <script>alert(1)</script>
  <defs><clipPath id="c"><rect width="4" height="4"/></clipPath></defs>
  <g clip-path="url(#c)" fill="url(#p)" stroke="currentColor"><path d="M0 0h4" onclick="x" id="p1"/></g>
  <foreignObject><div>html</div></foreignObject>
  <circle cx="2" cy="2" r="1"/>
</symbol>
<symbol id="no-prefix" viewBox="0 0 4 4"><path d="M0 0"/></symbol>
<symbol id="ico-ok" viewBox="0 0 9 9"><path d="M9 9"/></symbol>
</svg>`
	reg := map[string]iconGlyph{}
	require.NoError(t, loadIcons(reg, strings.NewReader(src)))

	require.Len(t, reg, 1, "only ico-prefixed symbols are icons")
	glyph := reg["ok"]
	assert.Equal(t, `<svg class="gno-icon" viewBox="0 0 4 4" fill="none"`, string(glyph.open))
	assert.Equal(t, `<g stroke="currentColor"><path d="M0 0h4"/></g><circle cx="2" cy="2" r="1"/></svg>`, string(glyph.body),
		"the first definition wins, and the allowlist drops title, script, defs, foreignObject, url() references, ids and handlers")
}

func TestParseIconTag(t *testing.T) {
	for _, tc := range []struct {
		line  string
		size  int
		name  string
		label string
	}{
		{line: `<gno-icon name="star" /> tail`, size: 24, name: "star"},
		{line: `<gno-icon name="star"/>`, size: 23, name: "star"},
		{line: `<gno-icon name=" star " label=" a > b " />`, size: 42, name: "star", label: "a > b"},
		{line: `<GNO-ICON NAME="star" />`, size: 24, name: "star"},
		{line: `<gno-icon />`, size: 12},
		{line: `<gno-icon name="star">`},
		{line: `</gno-icon>`},
		{line: `<gno-iconic name="star" />`},
		{line: `<gno-icon`},
		{line: `<gno-icon name="star`},
		{line: `<gno-ico`},
		{line: `<p>`},
	} {
		t.Run(tc.line, func(t *testing.T) {
			size, icon := parseIconTag([]byte(tc.line))
			assert.Equal(t, tc.size, size)
			if tc.size == 0 {
				assert.Nil(t, icon)
				return
			}
			require.NotNil(t, icon)
			assert.Equal(t, tc.name, icon.Name)
			assert.Equal(t, tc.label, icon.Label)
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
	assert.Contains(t, out, `<h2 id="plain-heading">Plain heading</h2>`)
	assert.NotContains(t, out, "gno-icon-name")

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
		"plain-heading=Plain heading",
	}, got, "an icon-only heading has no title, so the TOC drops it")
}

func TestStripIconTags(t *testing.T) {
	for in, want := range map[string]string{
		``:                                    ``,
		`Launch`:                              `Launch`,
		`<gno-icon name="rocket" /> Launch`:   ` Launch`,
		`a<gno-icon name="x"/>b<GNO-ICON />c`: `abc`,
		`a <b> c <gno-icon name="x">`:         `a <b> c <gno-icon name="x">`,
		`trailing <`:                          `trailing <`,
	} {
		assert.Equal(t, want, string(stripIconTags([]byte(in))), in)
	}
}

func TestIconInGFM(t *testing.T) {
	src := []byte(`| Status | Name |
|---|---|
| <gno-icon name="check-circle" label="ok" /> | alpha |
| <gno-icon name="unicorn" /> | beta |

~~<gno-icon name="trash" /> struck~~
`)
	var buf bytes.Buffer
	require.NoError(t, newProductionLikeMarkdown().Convert(src, &buf))
	out := buf.String()

	assert.Contains(t, out, `<td><svg class="gno-icon" viewBox="0 0 21 21" stroke-width="1.3" role="img" aria-label="ok">`)
	assert.Contains(t, out, `<td><!-- gno-icon: unknown name "unicorn" --></td>`)
	assert.Contains(t, out, `<del><svg class="gno-icon"`)
}

// TestIconCatalog keeps ICONS.md in sync with the registry. Regenerate it
// with `make icons` (or -update-golden-tests).
func TestIconCatalog(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`# gnoweb icons

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

An unknown name renders nothing (an HTML comment says why). To add an
icon, list it in ` + "`icons/icons.txt`" + ` and run ` + "`make icons`" + `.

`)
	names := slices.Sorted(maps.Keys(iconRegistry))
	fmt.Fprintf(&buf, "%d icons:\n\n| Icon | Name |\n|---|---|\n", len(names))
	for _, name := range names {
		fmt.Fprintf(&buf, "| <gno-icon name=\"%s\" /> | `%s` |\n", name, name)
	}

	if *update {
		require.NoError(t, os.WriteFile(iconCatalogFile, buf.Bytes(), 0o644))
	}
	want, err := os.ReadFile(iconCatalogFile)
	require.NoError(t, err)
	require.Equal(t, string(want), buf.String(), "%s is stale: run `make icons`", iconCatalogFile)
}

func BenchmarkIconParse(b *testing.B) {
	m := newProductionLikeMarkdown()
	for _, bc := range []struct{ name, line string }{
		{"icons", `Status <gno-icon name="check-circle" label="ok" /> and <gno-icon name="star" />. `},
		// The common case: '<' that is not an icon must stay on the cheap
		// prefix check.
		{"other-tags", `Some <b>bold</b> and <i>raw</i> html <br/> here. `},
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
