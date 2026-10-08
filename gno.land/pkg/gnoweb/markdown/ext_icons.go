// Package markdown — gno-icon extension.
//
// `<gno-icon name="star" />` renders a small inline icon anywhere inline
// markdown goes: headings, paragraphs, list items, table cells, link labels.
// An optional `label="…"` makes the icon meaningful (role="img" plus an
// aria-label); without it the icon is decorative (aria-hidden). No other
// attribute is read.
//
// Each icon is written inline as its own <svg>, so a page carries only the
// icons it uses. The glyphs come from iconRegistry, a Go table generated
// (`make icons`) from the chrome sprite components/ui/icons.html and the
// sets under icons/, through an element and attribute allowlist: nothing is
// parsed at run time, and a lookup is a map read.
//
// A tag is read by a bounded scanner (maxIconTagLen), on one line. A tag
// that is not self-closing, or that has a missing or unknown name, renders
// as an HTML comment saying so; a tag that never ends on its line is left
// to goldmark, which shows it as text.
package markdown

import (
	"bytes"
	"strconv"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const iconTagName = "gno-icon"

var iconTagPrefix = []byte("<" + iconTagName)

// iconGlyph is an iconRegistry entry: head holds the root <svg> attributes
// (viewBox, fill, stroke…) and body the shapes.
type iconGlyph struct {
	head, body string
}

// iconHeadStroke is the head every outline icon shares (the System UIcons
// set and the icons drawn for it): the generator writes it once here and
// strips the same values from each body, which inherit them.
const iconHeadStroke = `viewBox="0 0 21 21" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"`

// ----- AST node -----

// KindIcon is the node kind for Icon.
var KindIcon = ast.NewNodeKind("GnoIcon")

// Icon is the inline node for a `<gno-icon />` tag.
type Icon struct {
	ast.BaseInline
	iconTag
	// Source is where the tag sits in the document.
	Source text.Segment
}

// iconTag is what a tag says. Name and Label are the raw attribute values,
// aliasing the source. Kept apart from Icon so the scanner returns it on the
// stack and only a recognized tag allocates a node.
type iconTag struct {
	Name, Label []byte
	// SelfClosing is false for `<gno-icon name="…">`, which renders a hint
	// instead of the icon.
	SelfClosing bool
}

// Kind implements ast.Node.
func (*Icon) Kind() ast.NodeKind { return KindIcon }

// Dump implements ast.Node.
func (n *Icon) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"name":         string(n.Name),
		"label":        string(n.Label),
		"self_closing": strconv.FormatBool(n.SelfClosing),
	}, nil)
}

// ----- tag scanner -----

// maxIconTagLen bounds how far a tag is scanned. It keeps an unterminated
// `<gno-icon` from costing a scan to the end of the line, which, repeated
// on one line, made parsing quadratic; it also caps the label length.
const maxIconTagLen = 512

// parseIconTag reads a `<gno-icon …>` tag at the start of src, without
// allocating. It returns the tag's length, or 0 when src does not start with
// a gno-icon tag that ends (`/>` or `>`) on this line within maxIconTagLen
// bytes. Attribute names are case-insensitive and the first occurrence wins,
// as in HTML; values are returned raw, without entity decoding.
func parseIconTag(src []byte) (size int, icon iconTag) {
	n := len(iconTagPrefix)
	if len(src) <= n || !bytes.EqualFold(src[:n], iconTagPrefix) {
		return 0, icon
	}
	if c := src[n]; c != '/' && c != '>' && !util.IsSpace(c) {
		return 0, icon // e.g. <gno-iconic>
	}
	src = src[:min(len(src), maxIconTagLen)]
	if eol := bytes.IndexByte(src, '\n'); eol >= 0 {
		src = src[:eol] // a tag spans one line
	}

	var hasName, hasLabel bool
	for i := n; i < len(src); {
		switch c := src[i]; {
		case util.IsSpace(c):
			i++
			continue
		case c == '>':
			return i + 1, icon
		case c == '/':
			if i+1 < len(src) && src[i+1] == '>' {
				icon.SelfClosing = true
				return i + 2, icon
			}
			i++
			continue
		}

		// Attribute name, then an optional `= value`.
		start := i
		for i++; i < len(src) && !isIconAttrNameEnd(src[i]); i++ {
		}
		key := src[start:i]
		for i < len(src) && util.IsSpace(src[i]) {
			i++
		}
		var val []byte
		if i < len(src) && src[i] == '=' {
			for i++; i < len(src) && util.IsSpace(src[i]); i++ {
			}
			if i == len(src) {
				return 0, icon
			}
			if q := src[i]; q == '"' || q == '\'' {
				end := bytes.IndexByte(src[i+1:], q)
				if end < 0 {
					return 0, icon
				}
				val = src[i+1 : i+1+end]
				i += end + 2
			} else {
				start := i
				for i < len(src) && !util.IsSpace(src[i]) && src[i] != '>' {
					i++
				}
				val = src[start:i]
			}
		}

		switch {
		case !hasName && bytes.EqualFold(key, []byte("name")):
			icon.Name, hasName = bytes.TrimSpace(val), true
		case !hasLabel && bytes.EqualFold(key, []byte("label")):
			icon.Label, hasLabel = bytes.TrimSpace(val), true
		}
	}
	return 0, icon
}

func isIconAttrNameEnd(c byte) bool {
	return c == '/' || c == '>' || c == '=' || util.IsSpace(c)
}

// ----- parsers -----

type iconParser struct{}

var _ parser.InlineParser = (*iconParser)(nil)

// iconInHeadingKey is set on the parser context once an icon is parsed in
// a heading, so iconHeadingIDTransformer only walks documents that need it.
var iconInHeadingKey = parser.NewContextKey()

func (*iconParser) Trigger() []byte { return []byte{'<'} }

func (*iconParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, seg := block.PeekLine()
	size, tag := parseIconTag(line)
	if size == 0 {
		return nil
	}
	block.Advance(size)
	if parent.Kind() == ast.KindHeading {
		pc.Set(iconInHeadingKey, true)
	}
	return &Icon{iconTag: tag, Source: text.NewSegment(seg.Start, seg.Start+size)}
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
	if size, _ := parseIconTag(util.TrimLeftSpace(line)); size == 0 {
		return nil, parser.NoChildren
	}
	return p.BlockParser.Open(parent, reader, pc)
}

// iconHeadingIDTransformer rebuilds auto heading IDs without the icon tags.
// goldmark derives an ID from the heading's raw source line, so
// `## <gno-icon name="rocket" /> Launch` gets `gno-icon-namerocket-launch`;
// without the tag it gets `launch`, from the text the TOC shows. The tags
// removed are the Icon nodes goldmark parsed, so a tag shown as text (code
// span, backslash escape) stays, as any text does. Every heading is
// renumbered in document order with fresh IDs, so a `Launch` before or after
// the icon heading gets the suffix it would without icons. Other inline
// syntax is left in the ID, as goldmark leaves it.
type iconHeadingIDTransformer struct{}

func (*iconHeadingIDTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	if pc.Get(iconInHeadingKey) == nil {
		return
	}
	ids := parser.NewContext().IDs() // goldmark's generator, fresh
	src := reader.Source()
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}
		// Only auto IDs: without WithAutoHeadingID there is none to fix.
		if _, ok := h.AttributeString("id"); ok && h.Lines().Len() > 0 {
			line := h.Lines().At(h.Lines().Len() - 1) // the line goldmark uses
			h.SetAttributeString("id", ids.Generate(withoutIcons(h, line, src), ast.KindHeading))
		}
		return ast.WalkSkipChildren, nil
	})
}

// withoutIcons returns the source of line minus the Icons under n that sit
// on it, or the line itself when none does. The walk visits them in source
// order.
func withoutIcons(n ast.Node, line text.Segment, src []byte) []byte {
	var out []byte
	at := line.Start
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if icon, ok := c.(*Icon); ok && entering && icon.Source.Start >= at && icon.Source.Stop <= line.Stop {
			if out == nil {
				out = make([]byte, 0, line.Len())
			}
			out = append(out, src[at:icon.Source.Start]...)
			at = icon.Source.Stop
		}
		return ast.WalkContinue, nil
	})
	if out == nil {
		return src[line.Start:line.Stop]
	}
	return append(out, src[at:line.Stop]...)
}

// ----- renderer -----

type iconRenderer struct{}

func (*iconRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindIcon, renderIcon)
}

func renderIcon(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*Icon)

	glyph, ok := iconRegistry[string(n.Name)]
	switch {
	case !n.SelfClosing:
		w.WriteString(`<!-- gno-icon: write it self-closing, <gno-icon name="…" /> -->`)
		return ast.WalkContinue, nil
	case len(n.Name) == 0:
		w.WriteString("<!-- gno-icon: missing name -->")
		return ast.WalkContinue, nil
	case !ok:
		w.WriteString(`<!-- gno-icon: unknown name "`)
		gmhtml.DefaultWriter.RawWrite(w, n.Name)
		w.WriteString(`" -->`)
		return ast.WalkContinue, nil
	}

	w.WriteString(`<svg class="gno-icon" `)
	w.WriteString(glyph.head)
	if len(n.Label) > 0 {
		// The value is raw source: write it as goldmark writes text,
		// resolving character references and backslash escapes (`\|` in a
		// table cell), escaped, straight to w.
		w.WriteString(` role="img" aria-label="`)
		gmhtml.DefaultWriter.Write(w, n.Label)
		w.WriteString(`">`)
	} else {
		w.WriteString(` aria-hidden="true" focusable="false">`)
	}
	w.WriteString(glyph.body)
	w.WriteString("</svg>")

	// A decorative icon that is all a link or heading holds leaves it with
	// no accessible name: say so where the author will look.
	if len(n.Label) == 0 && aloneInNamedParent(n, source) {
		w.WriteString(`<!-- gno-icon: alone in a link or heading, add label="…" to name it -->`)
	}
	return ast.WalkContinue, nil
}

// aloneInNamedParent reports whether n is the only content, blank text
// aside, of a link or heading: an element whose accessible name comes from
// its content.
func aloneInNamedParent(n ast.Node, source []byte) bool {
	parent := n.Parent()
	if parent == nil {
		return false
	}
	switch parent.Kind() {
	case ast.KindLink, KindGnoLink, ast.KindHeading:
	default:
		return false
	}
	for c := parent.FirstChild(); c != nil; c = c.NextSibling() {
		if c == n {
			continue
		}
		t, ok := c.(*ast.Text)
		if !ok || !util.IsBlank(t.Segment.Value(source)) {
			return false
		}
	}
	return true
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
		parser.WithASTTransformers(
			util.Prioritized(&iconHeadingIDTransformer{}, 500),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&iconRenderer{}, 500),
	))
}
