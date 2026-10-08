package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindGnoPanel is the node kind of a `<gno-panel>` block.
var KindGnoPanel = ast.NewNodeKind("GnoPanel")

// GnoPanelNode is a framed block of ordinary markdown:
//
//	<gno-panel>
//	## Any markdown
//	</gno-panel>
//
// Standalone it renders as a highlighted panel; alone in a column the
// CSS turns it into a card. Like gno-columns, the tags are flat markers at
// document level; panelASTTransformer moves the blocks between them under
// the open marker.
type GnoPanelNode struct {
	ast.BaseBlock
	// tag is panelTagOpen (the container), panelTagClose or panelTagInvalid
	// (rendered as a comment).
	tag panelTagKind
}

// Kind implements ast.Node.
func (*GnoPanelNode) Kind() ast.NodeKind { return KindGnoPanel }

// Dump implements ast.Node.
func (n *GnoPanelNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// ----- tag recognition -----

type panelTagKind int

const (
	panelTagNone panelTagKind = iota
	panelTagOpen
	panelTagClose
	panelTagInvalid
)

var (
	panelOpenPrefix  = []byte("<gno-panel")
	panelClosePrefix = []byte("</gno-panel")

	// Prefixes of every gno-columns tag, separator included.
	columnsOpenPrefix  = []byte("<gno-columns")
	columnsClosePrefix = []byte("</gno-columns")
)

// parsePanelLineTag classifies a trimmed line. Only a bare `<gno-panel>`
// or `</gno-panel>` is valid: the attribute allowlist is empty, so any
// attribute, a self-closing form or trailing text makes the line invalid
// rather than silently ignored. Tag names are case-insensitive, like the
// other gno-* tags.
func parsePanelLineTag(line []byte) panelTagKind {
	kind, prefix := panelTagOpen, panelOpenPrefix
	if len(line) > 1 && line[1] == '/' {
		kind, prefix = panelTagClose, panelClosePrefix
	}
	attrs := 0
	size, selfClosing := scanGnoTag(line, prefix, len(line), func(_, _ []byte) { attrs++ })
	switch {
	case size == 0:
		return panelTagNone
	case size < len(line) && line[size] == '\r':
		// goldmark splits lines on '\n' only, so with bare-CR line endings
		// the whole input is one "line"; consuming it would drop the text.
		return panelTagNone
	case size != len(line) || selfClosing || attrs != 0:
		return panelTagInvalid
	}
	return kind
}

func hasPrefixFold(s, prefix []byte) bool {
	return len(s) >= len(prefix) && bytes.EqualFold(s[:len(prefix)], prefix)
}

func isColumnsTagLine(line []byte) bool {
	return (hasPrefixFold(line, columnsOpenPrefix) || hasPrefixFold(line, columnsClosePrefix)) &&
		parseLineTag(line) != GnoColumnTagUndefined
}

// ----- parse state -----

// panelOpenKey holds whether a panel is open at document level; it stays
// unset on a page without panel tags.
var panelOpenKey = parser.NewContextKey()

func panelOpen(pc parser.Context) bool {
	open, _ := pc.Get(panelOpenKey).(bool)
	return open
}

// ----- block parser -----

type panelParser struct{}

var _ parser.BlockParser = (*panelParser)(nil)

func (*panelParser) Trigger() []byte { return []byte{'<'} }

// Open reads a panel tag, at document level only (column content and a
// <gno-foreign> body count; see the ADR: this keeps sanitized user content
// from opening one after `> ` or `- `). A stray close tag, a malformed tag,
// a non-document parent, a panel in a panel or the depth cap yields an
// invalid leaf, never nil: nil would hand the line to the type-7 HTML block
// parser, which swallows every line up to the next blank one. A gno-columns
// tag ends an open panel and is left to the columns parser, which runs
// after this one.
func (*panelParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	line = trimTagLine(line)
	atDoc := parent.Kind() == ast.KindDocument

	kind := parsePanelLineTag(line)
	if kind == panelTagNone {
		if atDoc && panelOpen(pc) && isColumnsTagLine(line) {
			pc.Set(panelOpenKey, false)
			Pop(pc)
		}
		return nil, parser.NoChildren
	}

	reader.AdvanceToEOL()
	node := &GnoPanelNode{tag: panelTagInvalid}
	open := panelOpen(pc)
	switch {
	case !atDoc:
	case kind == panelTagOpen && !open && Push(pc): // the cap spans all gno-* blocks
		node.tag = panelTagOpen
		pc.Set(panelOpenKey, true)
	case kind == panelTagClose && open:
		node.tag = panelTagClose
		pc.Set(panelOpenKey, false)
		Pop(pc)
	}
	return node, parser.NoChildren
}

// Continue implements parser.BlockParser: markers are one line.
func (*panelParser) Continue(ast.Node, text.Reader, parser.Context) parser.State {
	return parser.Close
}

// Close implements parser.BlockParser.
func (*panelParser) Close(ast.Node, text.Reader, parser.Context) {}

// CanInterruptParagraph: true, like gno-columns.
func (*panelParser) CanInterruptParagraph() bool { return true }

// CanAcceptIndentedLine: false; 4+ spaces is an indented code block.
func (*panelParser) CanAcceptIndentedLine() bool { return false }

// panelHTMLBlockParser is goldmark's HTML block parser, except that while a
// panel is open, its close tag or a gno-columns tag also ends a
// document-level HTML block, which would otherwise run to the next blank
// line and swallow it.
type panelHTMLBlockParser struct{ parser.BlockParser }

func (p panelHTMLBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	if node.Parent().Kind() == ast.KindDocument && panelOpen(pc) {
		line, _ := reader.PeekLine()
		if tag := trimTagLine(line); parsePanelLineTag(tag) == panelTagClose || isColumnsTagLine(tag) {
			return parser.Close // the line reopens at document level
		}
	}
	return p.BlockParser.Continue(node, reader, pc)
}

// ----- AST transformer -----

type panelASTTransformer struct{}

// Transform moves the blocks after each open marker under it, up to its
// close marker (removed), the next gno-columns marker, or the end of the
// document; and pops the depth of a panel left open at EOF.
func (*panelASTTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	if pc.Get(panelOpenKey) == nil {
		return // no panel tag on this page
	}
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		panel, ok := n.(*GnoPanelNode)
		if !ok || panel.tag != panelTagOpen {
			continue
		}
		for c := panel.NextSibling(); c != nil; c = panel.NextSibling() {
			if m, ok := c.(*GnoPanelNode); ok && m.tag == panelTagClose {
				doc.RemoveChild(doc, m)
				break
			}
			if c.Kind() == KindGnoColumn {
				break
			}
			panel.AppendChild(panel, c)
		}
	}
	if panelOpen(pc) {
		pc.Set(panelOpenKey, false)
		Pop(pc)
	}
}

// ----- renderer -----

type panelRendererHTML struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (*panelRendererHTML) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGnoPanel, renderGnoPanel)
}

// renderGnoPanel writes constant markup only: no source byte reaches the
// wrapper, so the panel adds no injection path of its own.
func renderGnoPanel(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	switch {
	case node.(*GnoPanelNode).tag != panelTagOpen:
		if entering {
			w.WriteString("<!-- unexpected/invalid panel tag omitted -->\n")
		}
	case entering:
		w.WriteString("<section class=\"gno-panel\">\n")
	default:
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

// ----- extension registration -----

type panels struct{}

// ExtPanels is the singleton Extender for `<gno-panel>`.
var ExtPanels = &panels{}

// Extend registers the panel parser (ahead of the columns parser), the HTML
// block wrapper (ahead of goldmark's own), transformer and renderer.
func (*panels) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&panelParser{}, 499),
			util.Prioritized(panelHTMLBlockParser{parser.NewHTMLBlockParser()}, 899),
		),
		parser.WithASTTransformers(
			util.Prioritized(&panelASTTransformer{}, 500),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&panelRendererHTML{}, 500),
	))
}
