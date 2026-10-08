package markdown

import (
	"bytes"
	"fmt"
	"go/format"
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
// (iconset writes it); such an icon gets iconHeadStroke, and its body drops
// the values that head already sets: iconStrokeInherited, read from the
// constant so the two cannot drift.
var (
	iconStrokeRoot      = map[string]string{"viewbox": "0 0 21 21", "stroke-width": "1.3"}
	iconStrokeInherited = func() map[string]string {
		toks, _ := ParseHTMLTokens(strings.NewReader("<svg " + iconHeadStroke + ">"))
		m := map[string]string{}
		for _, a := range toks[0].Attr {
			if a.Key != "viewbox" {
				m[a.Key] = a.Val
			}
		}
		return m
	}()
)

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
		skipDepth int                 // > 0 inside an element dropped by the allowlist
		inherited []map[string]string // stroke icons: values set by enclosing elements
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
			var sets map[string]string
			if inherited != nil {
				sets = maps.Clone(inherited[len(inherited)-1])
			}
			body.WriteString("<" + tok.Data)
			for _, a := range tok.Attr {
				if !iconAttrs[a.Key] || strings.Contains(a.Val, "url(") {
					continue // url(#…) points at an element not copied
				}
				if sets != nil {
					if sets[a.Key] == a.Val {
						continue // inherited from the head or a parent
					}
					if _, ok := iconStrokeInherited[a.Key]; ok {
						sets[a.Key] = a.Val
					}
				}
				fmt.Fprintf(&body, ` %s="%s"`, a.Key, HTMLEscapeString(a.Val))
			}
			if tok.Type == html.SelfClosingTagToken {
				body.WriteString("/>")
			} else {
				body.WriteByte('>')
				if sets != nil {
					inherited = append(inherited, sets)
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
func iconHead(attrs []html.Attribute) (string, []map[string]string) {
	root := map[string]string{}
	for _, a := range attrs {
		if slices.Contains(iconRootAttrs, a.Key) && !strings.Contains(a.Val, "url(") {
			root[a.Key] = a.Val
		}
	}
	if maps.Equal(root, iconStrokeRoot) {
		return iconHeadStroke, []map[string]string{iconStrokeInherited}
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
		`<g fill="none" stroke="currentColor" stroke-linecap="round" transform="translate(3 3)">` +
		`<path d="M0 0" fill="none"/><circle r="1" fill="currentColor"/>` +
		`<g fill="currentColor"><path d="M1 1" fill="none"/><path d="M2 2" fill="currentColor"/></g>` +
		`</g></symbol>`
	table, err := buildIconTable([]byte(src))
	require.NoError(t, err)
	assert.Equal(t, iconGlyph{
		head: iconHeadStroke,
		body: `<g transform="translate(3 3)"><path d="M0 0"/><circle r="1" fill="currentColor"/>` +
			`<g fill="currentColor"><path d="M1 1" fill="none"/><path d="M2 2"/></g></g>`,
	}, table["s"], "a value is dropped only where the element would inherit it anyway")
}
