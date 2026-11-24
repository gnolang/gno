package markdown

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

// The icon table generator. iconRegistry (icons_gen.go) is built here from
// the icon sources, so nothing is parsed at run time; TestIconTable fails
// when the table is stale, and `make icons` (-update-golden-tests)
// regenerates it.

const iconTableFile = "icons_gen.go"

// iconSources are the symbol files the table is built from: the chrome
// sprite the page inlines, and the icons/ sets. A name defined twice, in one
// file or across files, is an error.
var iconSources = []struct {
	fsys fs.FS
	name string
}{
	{components.SharedPartialsFS(), "icons.html"},
	{os.DirFS("icons"), "drawn.svg"},
	{os.DirFS("icons"), "vendored.svg"},
}

// iconExcluded are chrome symbols kept out of the table: third-party brand
// marks, which a realm must not be able to wear, and the link badges, which
// <gno-foreign> content must not be able to add back (getLinkIcons drops
// them from foreign links).
var iconExcluded = map[string]bool{
	"github":   true,
	"twitter":  true,
	"discord":  true,
	"telegram": true,

	"external-link": true,
	"internal-link": true,
	"tx-link":       true,
	"user-link":     true,
}

// iconElements, iconAttrs and iconRootAttrs are the allowlist every glyph
// is re-serialized through. The tokenizer lowercases attribute names, so
// the root's viewbox is written back as viewBox.
var (
	iconElements = map[string]bool{
		"g": true, "path": true, "circle": true, "ellipse": true,
		"line": true, "polyline": true, "polygon": true, "rect": true,
	}
	iconAttrs = map[string]bool{
		"d": true, "cx": true, "cy": true, "r": true, "rx": true, "ry": true,
		"x": true, "y": true, "x1": true, "y1": true, "x2": true, "y2": true,
		"width": true, "height": true, "points": true, "transform": true,
		"fill": true, "fill-rule": true, "clip-rule": true, "opacity": true,
		"fill-opacity": true, "stroke": true, "stroke-width": true,
		"stroke-linecap": true, "stroke-linejoin": true,
		"stroke-miterlimit": true, "stroke-opacity": true,
	}
	iconRootAttrs = []string{"viewbox", "fill", "stroke", "stroke-width", "stroke-linecap", "stroke-linejoin"}
)

// iconStrokeRoot is the root an outline icon declares in its source file
// (iconset writes it); such an icon gets iconHeadStroke. Its shapes are
// painted with what they inherit, so the body is written against two sets
// of inherited values: the source's (SVG initial values under the source
// root) and the output's (iconHeadStroke's, read from the constant so the
// two cannot drift). A painting value is written wherever they differ: the
// head's stroke and round caps do not leak onto a shape the source leaves
// unstroked or butt-capped, and a value the head already sets is dropped.
var (
	iconStrokeRoot   = map[string]string{"viewbox": "0 0 21 21", "stroke-width": "1.3"}
	iconStrokeOutput = func() map[string]string {
		toks, _ := ParseHTMLTokens(strings.NewReader("<svg " + iconHeadStroke + ">"))
		m := map[string]string{}
		for _, a := range toks[0].Attr {
			if a.Key != "viewbox" {
				m[a.Key] = a.Val
			}
		}
		return m
	}()
	// iconStrokeSource is what the source root passes down: the SVG initial
	// values, with iconStrokeRoot's stroke width.
	iconStrokeSource = func() map[string]string {
		m := maps.Clone(svgInitialPaint)
		m["stroke-width"] = iconStrokeRoot["stroke-width"]
		return m
	}()
	// iconStrokeKeys is the order the painting values are written in.
	iconStrokeKeys = slices.Sorted(maps.Keys(iconStrokeOutput))
)

// iconPaint is what an element of an outline icon inherits, in the source
// and in the output.
type iconPaint struct{ src, out map[string]string }

// buildIconTable reads every `<symbol id="ico-NAME">` of the sources into
// glyphs, in source order.
func buildIconTable(sources ...[]byte) (map[string]iconGlyph, error) {
	table := map[string]iconGlyph{}
	for _, src := range sources {
		toks, err := ParseHTMLTokens(bytes.NewReader(src))
		if err != nil {
			return nil, err
		}
		if err := addIconSymbols(table, toks); err != nil {
			return nil, err
		}
	}
	// Excluded after reading, so a brand name defined twice still fails.
	for name := range iconExcluded {
		delete(table, name)
	}
	return table, nil
}

func addIconSymbols(table map[string]iconGlyph, toks []html.Token) error {
	var (
		name      string // symbol being read; "" outside a symbol
		head      string
		body      strings.Builder
		skipDepth int         // > 0 inside an element dropped by the allowlist
		inherited []iconPaint // outline icons: values set by enclosing elements
	)
	for _, tok := range toks {
		switch tok.Type {
		case html.StartTagToken, html.SelfClosingTagToken:
			if tok.Data == "symbol" {
				id, _ := ExtractAttr(tok.Attr, "id")
				var ok bool
				if name, ok = strings.CutPrefix(id, "ico-"); !ok {
					name = "" // not an icon symbol
				}
				head, inherited = iconHead(tok.Attr)
				body.Reset()
				continue
			}
			if name == "" {
				continue
			}
			if skipDepth > 0 || !iconElements[tok.Data] {
				if tok.Type == html.StartTagToken {
					skipDepth++
				}
				continue
			}
			var paint *iconPaint
			if inherited != nil {
				top := inherited[len(inherited)-1]
				paint = &iconPaint{maps.Clone(top.src), maps.Clone(top.out)}
			}
			body.WriteString("<" + tok.Data)
			for _, a := range tok.Attr {
				if !iconAttrs[a.Key] || strings.Contains(a.Val, "url(") {
					continue // url(#…) points at an element not copied
				}
				if _, ok := iconStrokeOutput[a.Key]; ok && paint != nil {
					paint.src[a.Key] = a.Val // written below if the output differs
					continue
				}
				fmt.Fprintf(&body, ` %s="%s"`, a.Key, HTMLEscapeString(a.Val))
			}
			if paint != nil {
				for _, key := range iconStrokeKeys {
					if key != "fill" && key != "stroke" && paint.src["stroke"] == "none" {
						continue // no stroke to shape; a stroked descendant writes it
					}
					if v := paint.src[key]; v != paint.out[key] {
						fmt.Fprintf(&body, ` %s="%s"`, key, HTMLEscapeString(v))
						paint.out[key] = v
					}
				}
			}
			if tok.Type == html.SelfClosingTagToken {
				body.WriteString("/>")
			} else {
				body.WriteByte('>')
				if paint != nil {
					inherited = append(inherited, *paint)
				}
			}

		case html.EndTagToken:
			switch {
			case name == "":
			case tok.Data == "symbol":
				if _, dup := table[name]; dup {
					return fmt.Errorf("icon %q defined twice", name)
				}
				table[name] = iconGlyph{head: head, body: body.String()}
				name = ""
			case skipDepth > 0:
				skipDepth--
			case iconElements[tok.Data]:
				body.WriteString("</" + tok.Data + ">")
				if len(inherited) > 1 {
					inherited = inherited[:len(inherited)-1]
				}
			}
		}
	}
	return nil
}

// iconHead returns the head for a symbol's attributes, and for an outline
// icon the values its shapes inherit from it.
func iconHead(attrs []html.Attribute) (string, []iconPaint) {
	root := map[string]string{}
	for _, a := range attrs {
		if slices.Contains(iconRootAttrs, a.Key) && !strings.Contains(a.Val, "url(") {
			root[a.Key] = a.Val
		}
	}
	if maps.Equal(root, iconStrokeRoot) {
		return iconHeadStroke, []iconPaint{{iconStrokeSource, iconStrokeOutput}}
	}
	var head []string
	for _, key := range iconRootAttrs {
		if v, ok := root[key]; ok {
			if key == "viewbox" {
				key = "viewBox"
			}
			head = append(head, fmt.Sprintf(`%s="%s"`, key, HTMLEscapeString(v)))
		}
	}
	return strings.Join(head, " "), nil
}

// iconTableSource renders table as icons_gen.go, one line per icon, sorted.
// An outline icon names the iconHeadStroke constant instead of repeating it.
func iconTableSource(table map[string]iconGlyph) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("// Code generated by `make icons` from components/ui/icons.html and icons/*.svg. DO NOT EDIT.\n\n")
	b.WriteString("package markdown\n\n")
	b.WriteString("// iconRegistry maps a <gno-icon> name to its glyph. See ext_icons_gen_test.go.\n")
	b.WriteString("var iconRegistry = map[string]iconGlyph{\n")
	for _, name := range slices.Sorted(maps.Keys(table)) {
		g := table[name]
		head := "`" + g.head + "`"
		if g.head == iconHeadStroke {
			head = "iconHeadStroke"
		}
		fmt.Fprintf(&b, "%q: {%s, `%s`},\n", name, head, g.body)
	}
	b.WriteString("}\n")
	return format.Source(b.Bytes())
}

// TestIconTable keeps icons_gen.go in sync with the icon sources.
func TestIconTable(t *testing.T) {
	var sources [][]byte
	for _, src := range iconSources {
		data, err := fs.ReadFile(src.fsys, src.name)
		require.NoError(t, err)
		sources = append(sources, data)
	}
	table, err := buildIconTable(sources...)
	require.NoError(t, err)
	got, err := iconTableSource(table)
	require.NoError(t, err)

	checkGenerated(t, iconTableFile, got)
}

// checkGenerated fails when file differs from got, or writes got to it with
// -update-golden-tests.
func checkGenerated(t *testing.T, file string, got []byte) {
	t.Helper()
	if *update {
		require.NoError(t, os.WriteFile(file, got, 0o644))
	}
	want, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Equal(t, string(want), string(got), "%s is stale: run `make icons`", file)
}

func TestBuildIconTableAllowlist(t *testing.T) {
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
</svg>`
	table, err := buildIconTable([]byte(src))
	require.NoError(t, err)

	require.Len(t, table, 1, "only ico-prefixed symbols are icons")
	assert.Equal(t, iconGlyph{
		head: `viewBox="0 0 4 4" fill="none"`,
		body: `<g stroke="currentColor"><path d="M0 0h4"/></g><circle cx="2" cy="2" r="1"/>`,
	}, table["ok"], "the allowlist drops title, script, defs, foreignObject, url() references, ids and handlers")
}

func TestBuildIconTableDuplicate(t *testing.T) {
	const a = `<symbol id="ico-x" viewBox="0 0 1 1"><path d="M0 0"/></symbol>`
	_, err := buildIconTable([]byte(a + a))
	require.ErrorContains(t, err, `"x" defined twice`)
	_, err = buildIconTable([]byte(a), []byte(a))
	require.ErrorContains(t, err, `"x" defined twice`, "across files too")
}

func TestBuildIconTableStrokeInheritance(t *testing.T) {
	const src = `<symbol id="ico-s" viewBox="0 0 21 21" stroke-width="1.3">` +
		`<g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" transform="translate(3 3)">` +
		`<path d="M0 0" fill="none"/><circle r="1" fill="currentColor"/>` +
		`<g fill="currentColor"><path d="M1 1" fill="none"/><path d="M2 2" fill="currentColor"/></g>` +
		`</g></symbol>` +
		`<symbol id="ico-dots" viewBox="0 0 21 21" stroke-width="1.3">` +
		`<g fill="currentColor"><circle r="1"/><circle r="2"/></g></symbol>` +
		`<symbol id="ico-butt" viewBox="0 0 21 21" stroke-width="1.3">` +
		`<g fill="none" stroke="currentColor"><path d="M0 0" stroke-linecap="round"/><path d="M1 1"/></g>` +
		`<path d="M2 2" fill="currentColor"/></symbol>`
	table, err := buildIconTable([]byte(src))
	require.NoError(t, err)
	assert.Equal(t, iconGlyph{
		head: iconHeadStroke,
		body: `<g transform="translate(3 3)"><path d="M0 0"/><circle r="1" fill="currentColor"/>` +
			`<g fill="currentColor"><path d="M1 1" fill="none"/><path d="M2 2"/></g></g>`,
	}, table["s"], "a value is dropped only where the element would inherit it anyway")
	assert.Equal(t, iconGlyph{
		head: iconHeadStroke,
		body: `<g fill="currentColor" stroke="none"><circle r="1"/><circle r="2"/></g>`,
	}, table["dots"], "a shape the source leaves unstroked stays unstroked")
	assert.Equal(t, iconGlyph{
		head: iconHeadStroke,
		body: `<g stroke-linecap="butt" stroke-linejoin="miter"><path d="M0 0" stroke-linecap="round"/><path d="M1 1"/></g>` +
			`<path d="M2 2" fill="currentColor" stroke="none"/>`,
	}, table["butt"], "the source's default caps and joins are kept")
}

// TestIconTableMatchesSource checks every icon of the icons/ sets paints its
// shapes as its source symbol does: the same fill and stroke, and on a
// stroked shape the same width, caps and joins. It resolves inheritance on
// both sides independently of the generator, so a head value the source
// never set (iconHeadStroke's stroke, round caps) shows up as a difference.
func TestIconTableMatchesSource(t *testing.T) {
	for _, file := range []string{"drawn.svg", "vendored.svg"} {
		data, err := os.ReadFile("icons/" + file)
		require.NoError(t, err)
		dec := xml.NewDecoder(bytes.NewReader(data))
		for {
			tok, err := dec.Token()
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)
			start, ok := tok.(xml.StartElement)
			if !ok || start.Name.Local != "symbol" {
				continue
			}
			var sym svgPaintNode
			require.NoError(t, dec.DecodeElement(&sym, &start))
			name := strings.TrimPrefix(sym.attr("id"), "ico-")
			g, ok := iconRegistry[name]
			if !ok {
				continue
			}
			var gen svgPaintNode
			require.NoError(t, xml.Unmarshal([]byte("<svg "+g.head+">"+g.body+"</svg>"), &gen))
			assert.Equal(t, sym.paints(svgInitialPaint), gen.paints(svgInitialPaint), "%s: icon %q", file, name)
		}
	}
}

// svgInitialPaint are the SVG initial values of the painting properties
// iconHeadStroke sets.
var svgInitialPaint = map[string]string{
	"fill": "black", "stroke": "none", "stroke-width": "1",
	"stroke-linecap": "butt", "stroke-linejoin": "miter",
}

type svgPaintNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr     `xml:",any,attr"`
	Children []svgPaintNode `xml:",any"`
}

func (n svgPaintNode) attr(key string) string {
	for _, a := range n.Attrs {
		if a.Name.Local == key {
			return a.Value
		}
	}
	return ""
}

// paints lists, in document order, how each shape under n is painted.
func (n svgPaintNode) paints(inherited map[string]string) []string {
	cur := maps.Clone(inherited)
	for key := range cur {
		if v := n.attr(key); v != "" {
			cur[key] = v
		}
	}
	if iconElements[n.XMLName.Local] && n.XMLName.Local != "g" {
		p := "fill=" + cur["fill"] + " stroke=" + cur["stroke"]
		if cur["stroke"] != "none" {
			p += fmt.Sprintf(" width=%s cap=%s join=%s", cur["stroke-width"], cur["stroke-linecap"], cur["stroke-linejoin"])
		}
		return []string{p}
	}
	var out []string
	for _, c := range n.Children {
		out = append(out, c.paints(cur)...)
	}
	return out
}
