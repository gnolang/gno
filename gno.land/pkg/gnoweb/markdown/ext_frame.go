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

// GnoFrameNode is a bordered block of ordinary markdown, which may hold
// complete gno-columns grids:
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
	// inner marks a frame opened in a column of a grid that is itself
	// inside a frame, and its close marker.
	inner bool
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
// attribute or a self-closing form makes the line invalid rather than
// silently ignored. Text after the tag makes it no tag line at all, so the
// text stays in a paragraph, as with gno-columns. Tag names are case-insensitive, like the
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
	case size < len(line):
		// Text after the tag: not a tag line, so the paragraph keeps the
		// text. With bare-CR line endings the "line" is the rest of the page.
		return frameTagNone
	case selfClosing || attrs != 0:
		return frameTagInvalid
	}
	return kind
}

// columnsLineTag returns the gno-columns tag on line, or
// GnoColumnTagUndefined. The prefix check spares the HTML tokenizer in
// parseLineTag on every other line.
func columnsLineTag(line []byte) GnoColumnTag {
	for _, prefix := range columnsTagPrefixes {
		if hasGnoTagPrefix(line, prefix) {
			return parseLineTag(line)
		}
	}
	return GnoColumnTagUndefined
}

// ----- parse state -----

// frameOpenKey holds whether a frame is open at document level; it stays
// unset on a page without frame tags.
var frameOpenKey = parser.NewContextKey()

func frameOpen(pc parser.Context) bool {
	open, _ := pc.Get(frameOpenKey).(bool)
	return open
}

// frameGridKey holds whether the gno-columns grid open now was opened
// inside the open frame.
var frameGridKey = parser.NewContextKey()

func frameGrid(pc parser.Context) bool {
	grid, _ := pc.Get(frameGridKey).(bool)
	return grid
}

// frameInnerKey holds whether a frame is open in a column of the open
// frame's grid.
var frameInnerKey = parser.NewContextKey()

func frameInner(pc parser.Context) bool {
	inner, _ := pc.Get(frameInnerKey).(bool)
	return inner
}

// frameRefusedKey holds whether a card opener was refused (an attribute,
// the depth cap) in a column of the open frame's grid; its close tag is
// then an invalid leaf, not the outer frame's close.
var frameRefusedKey = parser.NewContextKey()

func frameRefused(pc parser.Context) bool {
	refused, _ := pc.Get(frameRefusedKey).(bool)
	return refused
}

func gridOpen(pc parser.Context) bool {
	cctx, _ := pc.Get(columnContextKey).(*columnsContext)
	return cctx != nil && cctx.IsOpen
}

// endFrame marks the open frame as ended at parse time.
func endFrame(pc parser.Context) {
	pc.Set(frameOpenKey, false)
	pc.Set(frameGridKey, false)
	pc.Set(frameRefusedKey, false)
	Pop(pc)
}

// endInnerFrame marks the frame open in a column as ended at parse time.
func endInnerFrame(pc parser.Context) {
	pc.Set(frameInnerKey, false)
	Pop(pc)
}

// frameKeepsColumnsTag reports whether a gno-columns tag stays inside the
// open frame: one that opens a grid (with room left under the depth cap),
// a separator or close of a grid opened inside the frame, or a stray
// separator or close with no grid open (an invalid leaf, as outside a
// frame). Any other columns tag would leave a grid half inside, so it ends
// the frame.
func frameKeepsColumnsTag(tag GnoColumnTag, pc parser.Context) bool {
	gridOpen := gridOpen(pc)
	switch tag {
	case GnoColumnTagOpen:
		if gridOpen {
			// A second opener: an inert comment inside the frame's own
			// grid, otherwise a frame in a column reaching its grid.
			return frameGrid(pc)
		}
		if Get(pc) >= MaxGnoNestDepth {
			return false // the frame ends so the grid can open
		}
		pc.Set(frameGridKey, true)
		return true
	case GnoColumnTagSep:
		return !gridOpen || frameGrid(pc)
	case GnoColumnTagClose:
		if !gridOpen {
			return true
		}
		if !frameGrid(pc) {
			return false
		}
		pc.Set(frameGridKey, false)
		return true
	}
	return false
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
// parser, which swallows every line up to the next blank one. The one frame
// a frame holds is an inner one, in a column of its grid (a card); a close
// tag ends it first. A card refused there (an attribute, the depth cap)
// leaves its close tag an invalid leaf. A gno-columns tag is left to the
// columns parser, which runs after this one; it ends an inner frame, and
// ends the outer frame unless frameKeepsColumnsTag keeps it inside.
func (*frameParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	line = trimTagLine(line)
	atDoc := parent.Kind() == ast.KindDocument

	kind := parseFrameLineTag(line)
	if kind == frameTagNone {
		if atDoc && frameOpen(pc) {
			if tag := columnsLineTag(line); tag != GnoColumnTagUndefined {
				if frameInner(pc) {
					endInnerFrame(pc)
				}
				pc.Set(frameRefusedKey, false)
				if !frameKeepsColumnsTag(tag, pc) {
					endFrame(pc)
				}
			}
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
	case kind == frameTagOpen && frameGrid(pc) && gridOpen(pc) && !frameInner(pc) && Push(pc):
		node.tag, node.inner = frameTagOpen, true
		pc.Set(frameInnerKey, true)
	case kind == frameTagClose && frameInner(pc):
		node.tag, node.inner = frameTagClose, true
		endInnerFrame(pc)
	case kind == frameTagClose && frameRefused(pc):
		pc.Set(frameRefusedKey, false) // the refused card's close
	case kind == frameTagClose && open:
		node.tag = frameTagClose
		endFrame(pc)
	case line[1] != '/' && open && frameGrid(pc) && gridOpen(pc) && !frameInner(pc):
		pc.Set(frameRefusedKey, true) // a refused card opener
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
// frame is open, a frame tag or a gno-columns tag also ends a
// document-level HTML block, which would otherwise run to the next blank
// line and swallow it.
type frameHTMLBlockParser struct{ parser.BlockParser }

// Open leaves the line to goldmark's own HTML block parser while no frame
// is open, so a page without frames does not run the HTML block checks twice.
func (p frameHTMLBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	if !frameOpen(pc) {
		return nil, parser.NoChildren
	}
	return p.BlockParser.Open(parent, reader, pc)
}

func (p frameHTMLBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	if node.Parent().Kind() == ast.KindDocument && frameOpen(pc) {
		line, _ := reader.PeekLine()
		if tag := trimTagLine(line); parseFrameLineTag(tag) == frameTagClose || columnsLineTag(tag) != GnoColumnTagUndefined || (parseFrameLineTag(tag) != frameTagNone && frameGrid(pc) && gridOpen(pc)) {
			return parser.Close // the line reopens at document level
		}
	}
	return p.BlockParser.Continue(node, reader, pc)
}

// ----- AST transformer -----

type frameASTTransformer struct{}

// Transform moves the blocks after each open marker under it, up to its
// close marker (removed), a gno-columns marker that is not part of a grid
// closing inside the frame, or the end of the document; and pops the depth
// of the frames left open at EOF. It runs after the columns transformer,
// so a grid left open at EOF already has its close marker.
func (*frameASTTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	if pc.Get(frameOpenKey) == nil {
		return // no frame tag on this page
	}
	wrapFrames(doc)
	if frameInner(pc) {
		endInnerFrame(pc)
	}
	if frameOpen(pc) {
		endFrame(pc)
	}
}

// wrapFrames gives each open marker among parent's children its blocks. An
// outer frame then wraps the inner frames in its grids; an inner frame
// holds no grid, so it stops at the next columns marker.
func wrapFrames(parent ast.Node) {
	for n := parent.FirstChild(); n != nil; n = n.NextSibling() {
		frame, ok := n.(*GnoFrameNode)
		if !ok || frame.tag != frameTagOpen {
			continue
		}
		for c := frame.NextSibling(); c != nil; c = frame.NextSibling() {
			if m, ok := c.(*GnoFrameNode); ok && m.tag == frameTagClose && m.inner == frame.inner {
				parent.RemoveChild(parent, m)
				break
			}
			if c.Kind() == KindGnoColumn {
				col := c.(*GnoColumnNode)
				if !frame.inner && col.Tag == GnoColumnTagUndefined && col.inFrame {
					// A stray columns tag kept inside the frame.
					frame.AppendChild(frame, c)
					continue
				}
				end := gridCloseInFrame(col)
				if end == nil {
					// The frame ended at this tag while parsing, or the
					// grid closes after the frame.
					break
				}
				for c != end {
					c = c.NextSibling()
					frame.AppendChild(frame, c.PreviousSibling())
				}
			}
			frame.AppendChild(frame, c)
		}
		if !frame.inner {
			wrapFrames(frame)
		}
	}
}

// gridCloseInFrame returns the close marker of the grid that open opened
// inside the frame, when it comes before the frame's close marker; nil
// otherwise, and for any other columns marker. Inner frames are only in
// the frame's grids, so the first outer close marker met is the frame's
// own and the scan stops there: it reads the blocks the frame then takes,
// or the rest of a frame that ends at open.
func gridCloseInFrame(open *GnoColumnNode) ast.Node {
	if open.Tag != GnoColumnTagOpen || !open.inFrame {
		return nil
	}
	for n := open.NextSibling(); n != nil; n = n.NextSibling() {
		switch m := n.(type) {
		case *GnoFrameNode:
			if m.tag == frameTagClose && !m.inner {
				return nil
			}
		case *GnoColumnNode:
			if m.Tag == GnoColumnTagClose {
				return m
			}
		}
	}
	return nil
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
			// After the columns transformer (500), which closes a grid
			// left open at EOF.
			util.Prioritized(&frameASTTransformer{}, 501),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&frameRendererHTML{}, 500),
	))
}
