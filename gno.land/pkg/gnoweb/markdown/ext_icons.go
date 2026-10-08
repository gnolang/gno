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
// A tag is read by the shared bounded scanner (scanGnoTag), on one line. A tag
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
// set and the icons drawn for it): the generator writes it once here, and
// each body sets only the values that differ from it.
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
	// hintDone and hint memoize aloneInNamedParent for every icon of a
	// link or heading, so it walks that parent once.
	hintDone, hint bool
}

// iconTag is what a tag says. Name and Label are the raw attribute values,
// aliasing the source. Kept apart from Icon so parseIconTag returns it on the
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

// maxIconTagLen bounds how far a tag is scanned (see scanGnoTag): an
// unterminated `<gno-icon` costs at most this much, and it caps the label.
const maxIconTagLen = 512

// parseIconTag reads a `<gno-icon …>` tag at the start of src with the
// shared scanGnoTag, without allocating. It returns the tag's length, or 0.
// Attribute names are case-insensitive and the first occurrence wins, as in
// HTML; values are raw, without entity decoding. Unlike <gno-button>, a tag
// that is not self-closing is claimed too, so it renders a hint.
func parseIconTag(src []byte) (size int, icon iconTag) {
	var hasName, hasLabel bool
	size, icon.SelfClosing = scanGnoTag(src, iconTagPrefix, maxIconTagLen, func(key, val []byte) {
		switch {
		case !hasName && bytes.EqualFold(key, []byte("name")):
			icon.Name, hasName = bytes.TrimSpace(val), true
		case !hasLabel && bytes.EqualFold(key, []byte("label")):
			icon.Label, hasLabel = bytes.TrimSpace(val), true
		}
	})
	return size, icon
}

// ----- budget -----

// MaxIconsPerConvert caps the icons one Convert call renders, nested
// <gno-foreign> renders included. Each icon writes its own <svg> (up to a few
// KB), so without a cap 1 MiB of tags makes a page of over 100 MB. Tags past
// the cap fall through to goldmark's raw HTML path, as without this extension.
const MaxIconsPerConvert = 1000

type iconBudget struct {
	count int
}

// iconBudgetKey holds the *iconBudget shared by every parser context of one
// Convert call.
var iconBudgetKey = parser.NewContextKey()

// getIconBudget returns the icon budget of pc, creating it on first use.
func getIconBudget(pc parser.Context) *iconBudget {
	b, _ := pc.Get(iconBudgetKey).(*iconBudget)
	if b == nil {
		b = &iconBudget{}
		pc.Set(iconBudgetKey, b)
	}
	return b
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
	budget := getIconBudget(pc)
	if budget.count >= MaxIconsPerConvert {
		return nil
	}
	budget.count++
	block.Advance(size)
	if parent.Kind() == ast.KindHeading {
		pc.Set(iconInHeadingKey, true)
	}
	return &Icon{iconTag: tag, Source: text.NewSegment(seg.Start, seg.Start+size)}
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
	ids := newLinearIDs()
	src := reader.Source()
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}
		// Only auto IDs: without WithAutoHeadingID there is none to fix.
		if _, ok := h.AttributeString("id"); ok {
			var value []byte // an empty heading: goldmark's ID for no text
			if h.Lines().Len() > 0 {
				line := h.Lines().At(h.Lines().Len() - 1) // the line goldmark uses
				value = withoutIcons(h, line, src)
			}
			h.SetAttributeString("id", ids.Generate(value, ast.KindHeading))
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

// aloneInNamedParent reports whether n is the first icon of a link or
// heading (an element named by its content) that holds nothing else giving
// it a name: no text and no labeled icon, at any depth (`[*<icon/>*](…)`,
// `## <icon/><icon/>`). The hint is written once, on that first icon.
func aloneInNamedParent(n *Icon, source []byte) bool {
	parent := n.Parent()
	for parent != nil && parent.Type() == ast.TypeInline && !namedByContent(parent) {
		parent = parent.Parent() // emphasis and the like
	}
	if parent == nil || !namedByContent(parent) {
		return false
	}

	// The first icon to render walks the parent once and answers for every
	// icon in it, so a parent of k icons and any filler costs O(size).
	if n.hintDone {
		return n.hint
	}
	var first *Icon
	named := false
	_ = ast.Walk(parent, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch c := c.(type) {
		case *Icon:
			c.hintDone = true
			if first == nil {
				first = c
			}
			named = named || len(c.Label) > 0
		case *ast.Text:
			named = named || !util.IsBlank(c.Segment.Value(source))
		case *ast.String:
			named = named || !util.IsBlank(c.Value)
		}
		return ast.WalkContinue, nil
	})
	first.hint = !named
	return n.hint
}

func namedByContent(n ast.Node) bool {
	switch n.Kind() {
	case ast.KindLink, KindGnoLink, ast.KindHeading:
		return true
	}
	return false
}

// ----- extension -----

type iconExtension struct{}

// ExtIcons is the Goldmark extension for `<gno-icon />`.
var ExtIcons = &iconExtension{}

// Extend registers the icon parser just ahead of goldmark's raw-HTML inline
// parser (400), as <gno-button> does, and the shared line parser ahead of
// the HTML block parser (900).
func (e *iconExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&iconParser{}, 399),
		),
		parser.WithBlockParsers(
			util.Prioritized(newGnoTagLineParser(iconTagPrefix), 899),
		),
		parser.WithASTTransformers(
			util.Prioritized(&iconHeadingIDTransformer{}, 500),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&iconRenderer{}, 500),
	))
}
