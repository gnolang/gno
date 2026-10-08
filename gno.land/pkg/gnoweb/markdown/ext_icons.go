// Package markdown — gno-icon extension.
//
// `<gno-icon name="star" />` renders a small inline icon anywhere inline
// markdown goes: headings, paragraphs, list items, table cells, link labels.
// An optional `label="…"` makes the icon meaningful (role="img" plus an
// aria-label); without it the icon is decorative (aria-hidden). No other
// attribute is read.
//
// Each icon is written inline as its own <svg>, so a page carries only the
// icons it uses. The glyphs come from one registry built at init from the
// same symbols the chrome uses (components/ui/icons.html) plus the vendored
// and hand-drawn sets under icons/. Every glyph is re-serialized through an
// element and attribute allowlist, so whatever the source files hold, the
// output is plain shape markup.
//
// Only a well-formed self-closing tag is claimed. A tag that is not
// self-closing falls through to goldmark's raw-HTML parser, which safe mode
// strips; a self-closing tag with a missing or unknown name renders as an
// HTML comment saying so.
package markdown

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

const iconTagName = "gno-icon"

var iconTagPrefix = []byte("<" + iconTagName)

//go:embed icons/*.svg
var iconFiles embed.FS

// iconSources lists the symbol files the registry reads. On a name present
// in several files the first one wins, as it would for an in-page `#ico-…`
// reference; TestIconRegistryNoShadowing keeps the files from colliding.
var iconSources = []iconSource{
	{components.SharedPartialsFS(), "icons.html"},
	{iconFiles, "icons/drawn.svg"},
	{iconFiles, "icons/vendored.svg"},
}

// iconExcluded are chrome symbols that stay out of the registry: third-party
// brand marks, which a realm must not be able to wear.
var iconExcluded = map[string]bool{
	"github":   true,
	"twitter":  true,
	"discord":  true,
	"telegram": true,
}

// iconElements, iconAttrs and iconRootAttrs are the allowlist a glyph is
// re-serialized through. The tokenizer lowercases attribute names, so the
// root's viewbox is written back as viewBox.
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
	iconRootAttrs = map[string]bool{
		"viewbox": true, "fill": true, "stroke": true, "stroke-width": true,
		"stroke-linecap": true, "stroke-linejoin": true,
	}
)

// iconGlyph is a registry entry, pre-rendered so the renderer only copies
// bytes: open is `<svg class="gno-icon" viewBox=… …` (unterminated, the
// renderer appends the accessibility attributes) and body is the shapes plus
// `</svg>`.
type iconGlyph struct {
	open, body []byte
}

var iconRegistry = mustLoadIcons()

type iconSource struct {
	fsys fs.FS
	name string
}

func mustLoadIcons() map[string]iconGlyph {
	reg := map[string]iconGlyph{}
	for _, src := range iconSources {
		if err := loadIconSource(reg, src); err != nil {
			panic(fmt.Sprintf("gno-icon: %s: %s", src.name, err))
		}
	}
	return reg
}

func loadIconSource(reg map[string]iconGlyph, src iconSource) error {
	f, err := src.fsys.Open(src.name)
	if err != nil {
		return err
	}
	defer f.Close()
	return loadIcons(reg, f)
}

// loadIcons adds every `<symbol id="ico-NAME">` of r to reg, keeping an
// existing entry on a duplicate name.
func loadIcons(reg map[string]iconGlyph, r io.Reader) error {
	toks, err := ParseHTMLTokens(r)
	if err != nil {
		return err
	}

	var (
		name      string // symbol being read; "" outside a symbol
		open      bytes.Buffer
		body      bytes.Buffer
		skipDepth int // > 0 inside an element dropped by the allowlist
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
				open.Reset()
				body.Reset()
				open.WriteString(`<svg class="gno-icon"`)
				writeIconAttrs(&open, tok.Attr, iconRootAttrs)
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
			body.WriteString("<" + tok.Data)
			writeIconAttrs(&body, tok.Attr, iconAttrs)
			if tok.Type == html.SelfClosingTagToken {
				body.WriteString("/>")
			} else {
				body.WriteByte('>')
			}

		case html.EndTagToken:
			switch {
			case name == "":
			case tok.Data == "symbol":
				if _, dup := reg[name]; !dup && !iconExcluded[name] {
					body.WriteString("</svg>")
					reg[name] = iconGlyph{
						open: bytes.Clone(open.Bytes()),
						body: bytes.Clone(body.Bytes()),
					}
				}
				name = ""
			case skipDepth > 0:
				skipDepth--
			case iconElements[tok.Data]:
				body.WriteString("</" + tok.Data + ">")
			}
		}
	}
	return nil
}

// writeIconAttrs writes the allowed attributes of attrs, escaped. A value
// referencing another element (url(#…)) is dropped: the referenced element
// is not copied, and an inline copy per use would duplicate its id.
func writeIconAttrs(buf *bytes.Buffer, attrs []html.Attribute, allowed map[string]bool) {
	for _, a := range attrs {
		if !allowed[a.Key] || strings.Contains(a.Val, "url(") {
			continue
		}
		key := a.Key
		if key == "viewbox" {
			key = "viewBox"
		}
		fmt.Fprintf(buf, ` %s="%s"`, key, HTMLEscapeString(a.Val))
	}
}

// ----- AST node -----

// KindIcon is the node kind for Icon.
var KindIcon = ast.NewNodeKind("GnoIcon")

// Icon is the inline node for a `<gno-icon />` tag.
type Icon struct {
	ast.BaseInline
	Name  string
	Label string
}

// Kind implements ast.Node.
func (*Icon) Kind() ast.NodeKind { return KindIcon }

// Dump implements ast.Node.
func (n *Icon) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"name":  n.Name,
		"label": n.Label,
	}, nil)
}

// parseIconTag reads a `<gno-icon … />` tag at the start of line. It returns
// the tag's length in bytes and the node, or 0 and nil when line does not
// start with a well-formed self-closing gno-icon tag. The tokenizer, which
// handles quoting (a `>` inside a label), only runs once the prefix matched.
func parseIconTag(line []byte) (int, *Icon) {
	n := len(iconTagPrefix)
	if len(line) <= n || !bytes.EqualFold(line[:n], iconTagPrefix) {
		return 0, nil
	}
	if c := line[n]; c != '/' && !util.IsSpace(c) {
		return 0, nil // e.g. <gno-iconic>
	}

	z := html.NewTokenizer(bytes.NewReader(line))
	if z.Next() != html.SelfClosingTagToken {
		return 0, nil
	}
	size := len(z.Raw())
	tok := z.Token()

	name, _ := ExtractAttr(tok.Attr, "name")
	label, _ := ExtractAttr(tok.Attr, "label")
	return size, &Icon{
		Name:  strings.TrimSpace(name),
		Label: strings.TrimSpace(label),
	}
}

// ----- parsers -----

type iconParser struct{}

var _ parser.InlineParser = (*iconParser)(nil)

func (*iconParser) Trigger() []byte { return []byte{'<'} }

func (*iconParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, _ := block.PeekLine()
	size, icon := parseIconTag(line)
	if icon == nil {
		return nil
	}
	block.Advance(size)
	return icon
}

// iconParagraphParser opens a paragraph on a line that starts with a
// gno-icon tag. Without it, a line holding only `<gno-icon … />` is a
// CommonMark type-7 HTML block, which safe mode strips. It delegates to
// goldmark's own paragraph parser, so the paragraph behaves like any other.
type iconParagraphParser struct {
	parser.BlockParser
}

func (*iconParagraphParser) Trigger() []byte { return []byte{'<'} }

func (p *iconParagraphParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if _, icon := parseIconTag(util.TrimLeftSpace(line)); icon == nil {
		return nil, parser.NoChildren
	}
	return p.BlockParser.Open(parent, reader, pc)
}

// iconIDs generates auto heading IDs from the heading's source line with
// gno-icon tags removed. goldmark derives the ID from the raw line, so
// `## <gno-icon name="rocket" /> Launch` would get `gno-icon-namerocket-launch`;
// stripped, it gets `launch`, from the same text the TOC shows. Generating
// at the usual time, rather than fixing IDs after parsing, keeps duplicate
// numbering in document order. Installed by NewGnoParserContext.
//
// Deliberately narrow: other inline syntax (links, emphasis) also reaches
// goldmark's IDs through the raw line, and stripping it all would move
// existing anchors. Only the icon tag, which is new, is taken out.
type iconIDs struct {
	parser.IDs
}

func (ids iconIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	return ids.IDs.Generate(stripIconTags(value), kind)
}

// stripIconTags returns value without its gno-icon tags. It returns value
// itself, without allocating, when there is no '<' to look at.
func stripIconTags(value []byte) []byte {
	if bytes.IndexByte(value, '<') < 0 {
		return value
	}
	out := make([]byte, 0, len(value))
	for {
		i := bytes.IndexByte(value, '<')
		if i < 0 {
			return append(out, value...)
		}
		out = append(out, value[:i]...)
		if size, icon := parseIconTag(value[i:]); icon != nil {
			value = value[i+size:]
			continue
		}
		out = append(out, '<')
		value = value[i+1:]
	}
}

// ----- renderer -----

type iconRenderer struct{}

func (*iconRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindIcon, renderIcon)
}

func renderIcon(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n, ok := node.(*Icon)
	if !entering || !ok {
		return ast.WalkContinue, nil
	}

	glyph, ok := iconRegistry[n.Name]
	switch {
	case n.Name == "":
		w.WriteString("<!-- gno-icon: missing name -->")
		return ast.WalkContinue, nil
	case !ok:
		fmt.Fprintf(w, `<!-- gno-icon: unknown name "%s" -->`, HTMLEscapeString(n.Name))
		return ast.WalkContinue, nil
	}

	w.Write(glyph.open)
	if n.Label != "" {
		fmt.Fprintf(w, ` role="img" aria-label="%s">`, HTMLEscapeString(n.Label))
	} else {
		w.WriteString(` aria-hidden="true">`)
	}
	w.Write(glyph.body)
	return ast.WalkContinue, nil
}

// ----- extension -----

type iconExtension struct{}

// ExtIcons is the Goldmark extension for `<gno-icon />`.
var ExtIcons = &iconExtension{}

// Extend registers the icon parsers ahead of goldmark's autolink (300) and
// raw-HTML (400) inline parsers, and the line parser ahead of the HTML block
// parser (900).
func (e *iconExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&iconParser{}, 250),
		),
		parser.WithBlockParsers(
			util.Prioritized(&iconParagraphParser{parser.NewParagraphParser()}, 899),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&iconRenderer{}, 500),
	))
}
