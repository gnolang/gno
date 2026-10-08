package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindGnoFrame is the node kind of a `<gno-frame>` block.
var KindGnoFrame = ast.NewNodeKind("GnoFrame")

// GnoFrameNode is a bordered block of ordinary markdown:
//
//	<gno-frame>
//	## Any markdown
//	</gno-frame>
//
// Standalone it renders as a bordered block; alone in a column the
// CSS turns it into a card. Like gno-columns, the tags are flat markers at
// document level; frameASTTransformer moves the blocks between them under
// the open marker.
type GnoFrameNode struct {
	ast.BaseBlock
	// tag is frameTagOpen (the container), frameTagClose or frameTagInvalid
	// (rendered as a comment).
	tag frameTagKind
}

// Kind implements ast.Node.
func (*GnoFrameNode) Kind() ast.NodeKind { return KindGnoFrame }

// Dump implements ast.Node.
func (n *GnoFrameNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// ----- tag recognition -----

type frameTagKind int

const (
	frameTagNone frameTagKind = iota
	frameTagOpen
	frameTagClose
	frameTagInvalid
)

var (
	frameOpenPrefix  = []byte("<gno-frame")
	frameClosePrefix = []byte("</gno-frame")

	// Every gno-columns tag name parseLineTag (ext_columns.go) accepts; a new
	// columns tag must be added here too.
	columnsTagPrefixes = [][]byte{[]byte("<gno-columns"), []byte("<gno-columns-sep"), []byte("</gno-columns")}
)

// parseFrameLineTag classifies a trimmed line. Only a bare `<gno-frame>`
// or `</gno-frame>` is valid: the attribute allowlist is empty, so any
// attribute, a self-closing form or trailing text makes the line invalid
// rather than silently ignored. Tag names are case-insensitive, like the
// other gno-* tags.
func parseFrameLineTag(line []byte) frameTagKind {
	kind, prefix := frameTagOpen, frameOpenPrefix
	if len(line) > 1 && line[1] == '/' {
		kind, prefix = frameTagClose, frameClosePrefix
	}
	attrs := 0
	size, selfClosing := scanGnoTag(line, prefix, len(line), func(_, _ []byte) { attrs++ })
	switch {
	case size == 0:
		return frameTagNone
	case size < len(line) && line[size] == '\r':
		// goldmark splits lines on '\n' only, so with bare-CR line endings
		// the whole input is one "line"; consuming it would drop the text.
		return frameTagNone
	case size != len(line) || selfClosing || attrs != 0:
		return frameTagInvalid
	}
	return kind
}

// isColumnsTagLine reports whether line is a gno-columns tag. The prefix
// check spares the HTML tokenizer in parseLineTag on every other line.
func isColumnsTagLine(line []byte) bool {
	for _, prefix := range columnsTagPrefixes {
		if hasGnoTagPrefix(line, prefix) {
			return parseLineTag(line) != GnoColumnTagUndefined
		}
	}
	return false
}

// ----- parse state -----

// frameOpenKey holds whether a frame is open at document level; it stays
// unset on a page without frame tags.
var frameOpenKey = parser.NewContextKey()

func frameOpen(pc parser.Context) bool {
	open, _ := pc.Get(frameOpenKey).(bool)
	return open
}

// ----- block parser -----

type frameParser struct{}

var _ parser.BlockParser = (*frameParser)(nil)

func (*frameParser) Trigger() []byte { return []byte{'<'} }

// Open reads a frame tag, at document level only (column content and a
// <gno-foreign> body count; see the ADR: this keeps sanitized user content
// from opening one after `> ` or `- `). A stray close tag, a malformed tag,
// a non-document parent, a frame in a frame or the depth cap yields an
// invalid leaf, never nil: nil would hand the line to the type-7 HTML block
// parser, which swallows every line up to the next blank one. A gno-columns
// tag ends an open frame and is left to the columns parser, which runs
// after this one.
func (*frameParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	line = trimTagLine(line)
	atDoc := parent.Kind() == ast.KindDocument

	kind := parseFrameLineTag(line)
	if kind == frameTagNone {
		if atDoc && frameOpen(pc) && isColumnsTagLine(line) {
			pc.Set(frameOpenKey, false)
			Pop(pc)
		}
		return nil, parser.NoChildren
	}

	reader.AdvanceToEOL()
	node := &GnoFrameNode{tag: frameTagInvalid}
	open := frameOpen(pc)
	switch {
	case !atDoc:
	case kind == frameTagOpen && !open && Push(pc): // the cap spans all gno-* blocks
		node.tag = frameTagOpen
		pc.Set(frameOpenKey, true)
	case kind == frameTagClose && open:
		node.tag = frameTagClose
		pc.Set(frameOpenKey, false)
		Pop(pc)
	}
	return node, parser.NoChildren
}

// Continue implements parser.BlockParser: markers are one line.
func (*frameParser) Continue(ast.Node, text.Reader, parser.Context) parser.State {
	return parser.Close
}

// Close implements parser.BlockParser.
func (*frameParser) Close(ast.Node, text.Reader, parser.Context) {}

// CanInterruptParagraph: true, like gno-columns.
func (*frameParser) CanInterruptParagraph() bool { return true }

// CanAcceptIndentedLine: false; 4+ spaces is an indented code block.
func (*frameParser) CanAcceptIndentedLine() bool { return false }

// frameHTMLBlockParser is goldmark's HTML block parser, except that while a
// frame is open, its close tag or a gno-columns tag also ends a
// document-level HTML block, which would otherwise run to the next blank
// line and swallow it.
type frameHTMLBlockParser struct{ parser.BlockParser }

func (p frameHTMLBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	if node.Parent().Kind() == ast.KindDocument && frameOpen(pc) {
		line, _ := reader.PeekLine()
		if tag := trimTagLine(line); parseFrameLineTag(tag) == frameTagClose || isColumnsTagLine(tag) {
			return parser.Close // the line reopens at document level
		}
	}
	return p.BlockParser.Continue(node, reader, pc)
}

// ----- AST transformer -----

type frameASTTransformer struct{}

// Transform moves the blocks after each open marker under it, up to its
// close marker (removed), the next gno-columns marker, or the end of the
// document; and pops the depth of a frame left open at EOF.
func (*frameASTTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	if pc.Get(frameOpenKey) == nil {
		return // no frame tag on this page
	}
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		frame, ok := n.(*GnoFrameNode)
		if !ok || frame.tag != frameTagOpen {
			continue
		}
		for c := frame.NextSibling(); c != nil; c = frame.NextSibling() {
			if m, ok := c.(*GnoFrameNode); ok && m.tag == frameTagClose {
				doc.RemoveChild(doc, m)
				break
			}
			if c.Kind() == KindGnoColumn {
				break
			}
			frame.AppendChild(frame, c)
		}
	}
	if frameOpen(pc) {
		pc.Set(frameOpenKey, false)
		Pop(pc)
	}
}

// ----- renderer -----

type frameRendererHTML struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (*frameRendererHTML) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGnoFrame, renderGnoFrame)
}

// renderGnoFrame writes constant markup only: no source byte reaches the
// wrapper, so the frame adds no injection path of its own.
func renderGnoFrame(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	switch {
	case node.(*GnoFrameNode).tag != frameTagOpen:
		if entering {
			w.WriteString("<!-- unexpected/invalid frame tag omitted -->\n")
		}
	case entering:
		w.WriteString("<section class=\"gno-frame\">\n")
	default:
		w.WriteString("</section>\n")
	}
	return ast.WalkContinue, nil
}

// ----- extension registration -----

type frames struct{}

// ExtFrames is the singleton Extender for `<gno-frame>`.
var ExtFrames = &frames{}

// Extend registers the frame parser (ahead of the columns parser), the HTML
// block wrapper (ahead of goldmark's own), transformer and renderer.
func (*frames) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&frameParser{}, 499),
			util.Prioritized(frameHTMLBlockParser{parser.NewHTMLBlockParser()}, 899),
		),
		parser.WithASTTransformers(
			util.Prioritized(&frameASTTransformer{}, 500),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&frameRendererHTML{}, 500),
	))
}
