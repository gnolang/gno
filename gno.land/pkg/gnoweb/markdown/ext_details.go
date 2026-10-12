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

// KindGnoDetails is the node kind of a `<gno-details>` block.
var KindGnoDetails = ast.NewNodeKind("GnoDetails")

// GnoDetailsNode is a neutral collapsible block, rendered as a bare
// <details>/<summary> pair without the chrome of an alert:
//
//	<gno-details summary="Logs" open>
//	arbitrary **markdown**
//	</gno-details>
//
// Like gno-frame, the tags are one-line markers; detailsASTTransformer moves
// the blocks between an open marker and its close marker under the open
// one. Each line therefore belongs to the innermost block that takes it: a
// close tag inside a code fence or a <gno-foreign> body never reaches this
// parser.
type GnoDetailsNode struct {
	ast.BaseBlock
	// tag is detailsTagOpen (the container), detailsTagClose or
	// detailsTagInvalid (rendered as a comment).
	tag detailsTagKind
	// Open renders the block expanded.
	Open bool
	// Summary is the raw `summary` attribute, escaped when rendered.
	Summary []byte
	// close is an open marker's close marker, set when the close tag pairs
	// with it during parsing.
	close *GnoDetailsNode
	// refused marks an opener turned down, by the depth cap or an invalid
	// attribute; it stays on the parse stack so its own close is the one it
	// takes.
	refused bool
}

// Kind implements ast.Node.
func (*GnoDetailsNode) Kind() ast.NodeKind { return KindGnoDetails }

// Dump implements ast.Node.
func (n *GnoDetailsNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// ----- tag recognition -----

type detailsTagKind int

const (
	detailsTagNone detailsTagKind = iota
	detailsTagOpen
	detailsTagClose
	detailsTagInvalid
)

var (
	detailsOpenPrefix  = []byte("<gno-details")
	detailsClosePrefix = []byte("</gno-details")
)

// parseDetailsLineTag classifies a trimmed line and returns an opener's
// attributes. The opener takes `summary="…"` and a bare `open`; any other
// attribute, a valued `open`, a repeated attribute, an attribute on the
// close tag or a self-closing form makes the tag invalid rather than
// silently ignored, valid reporting false. Text after the tag makes it no
// tag line at all, so the text stays in a paragraph, as with gno-frame. Tag
// names are case-insensitive, like the other gno-* tags.
func parseDetailsLineTag(line []byte) (kind detailsTagKind, valid, open bool, summary []byte) {
	kind, prefix := detailsTagOpen, detailsOpenPrefix
	if len(line) > 1 && line[1] == '/' {
		kind, prefix = detailsTagClose, detailsClosePrefix
	}
	valid, seenOpen, seenSummary := true, false, false
	size, selfClosing := scanGnoTag(line, prefix, len(line), func(key, val []byte) {
		switch {
		case kind == detailsTagOpen && bytes.EqualFold(key, []byte("open")) && val == nil && !seenOpen:
			seenOpen = true
		case kind == detailsTagOpen && bytes.EqualFold(key, []byte("summary")) && !seenSummary:
			seenSummary, summary = true, val
		default:
			valid = false
		}
	})
	switch {
	case size == 0, size < len(line):
		return detailsTagNone, false, false, nil
	case selfClosing:
		return kind, false, false, nil
	}
	return kind, valid, seenOpen, summary
}

// ----- parse state -----

// detailsEntry is an opener not yet closed and the container it sits in.
type detailsEntry struct {
	node   *GnoDetailsNode
	parent ast.Node
}

// detailsStackKey holds the openers not yet closed, refused ones included,
// innermost last. Each entry's container holds the one above it, since
// releaseDetails drops an entry once its container ends.
var detailsStackKey = parser.NewContextKey()

func detailsStack(pc parser.Context) []detailsEntry {
	stack, _ := pc.Get(detailsStackKey).([]detailsEntry)
	return stack
}

// endDetails drops the entries from i up and gives back their levels.
func endDetails(pc parser.Context, stack []detailsEntry, i int) {
	for _, e := range stack[i:] {
		if !e.node.refused {
			Pop(pc)
		}
	}
	pc.Set(detailsStackKey, stack[:i])
}

// releaseDetails ends the openers whose container a block opening under
// parent no longer sits in: a block left unclosed in a list item or a
// blockquote ends with it. Goldmark opens a line's new blocks before it
// closes the containers that line ends, so the parent it hands Open is the
// one measure of which containers are still open.
func releaseDetails(pc parser.Context, parent ast.Node) {
	stack := detailsStack(pc)
	i := len(stack)
	for i > 0 && !holds(stack[i-1].parent, parent) {
		i--
	}
	if i < len(stack) {
		endDetails(pc, stack, i)
	}
}

// holds reports whether c is n or one of its ancestors.
func holds(c, n ast.Node) bool {
	for ; n != nil; n = n.Parent() {
		if n == c {
			return true
		}
	}
	return false
}

// ----- block parser -----

type detailsParser struct{}

var _ parser.BlockParser = (*detailsParser)(nil)

// Trigger takes `>` as well as `<`: every gno-* block that takes a depth
// level opens on one of the two, so Open releases the ended details blocks
// before any of them reads the depth.
func (*detailsParser) Trigger() []byte { return []byte{'<', '>'} }

// Open reads a details tag at any level, so a block may sit in a list item,
// a blockquote, an alert or another details block. An opener takes a level
// of the gno-* depth cap and is refused past it, or for an invalid
// attribute; a close tag pairs with the innermost opener, when that opener
// sits in the same container, and gives its level back. A refused opener,
// its close, a stray close tag and a malformed tag each render as a
// comment. Every tag line yields a leaf, never nil: nil would hand the line
// to the type-7 HTML block parser, which swallows every line up to the
// next blank one. A gno-columns or gno-frame tag at document level ends
// every open details block, as the transformer stops them there.
func (*detailsParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	releaseDetails(pc, parent)
	line, _ := reader.PeekLine()
	// Goldmark never offers a line indented 4 columns or more; a tab after a
	// container prefix counts to its next stop.
	_, pos := util.IndentWidth(line, reader.LineOffset())
	tag := util.TrimRightSpace(line[pos:])
	kind, valid, open, summary := parseDetailsLineTag(tag)
	if kind == detailsTagNone {
		if parent.Kind() == ast.KindDocument && (columnsLineTag(tag) != GnoColumnTagUndefined || parseFrameLineTag(tag) != frameTagNone) {
			// Released above, every opener left sits at document level.
			endDetails(pc, detailsStack(pc), 0)
		}
		return nil, parser.NoChildren
	}
	reader.AdvanceToEOL()

	stack := detailsStack(pc)
	node := &GnoDetailsNode{tag: detailsTagInvalid}
	switch {
	case kind == detailsTagOpen:
		if valid && Push(pc) { // the cap spans all gno-* blocks
			node.tag, node.Open, node.Summary = detailsTagOpen, open, summary
		} else {
			node.refused = true
		}
		pc.Set(detailsStackKey, append(stack, detailsEntry{node, parent}))
	case valid && len(stack) > 0 && stack[len(stack)-1].parent == parent:
		top := len(stack) - 1
		node.tag = detailsTagClose
		stack[top].node.close = node
		endDetails(pc, stack, top)
	}
	return node, parser.NoChildren
}

// Continue implements parser.BlockParser: markers are one line.
func (*detailsParser) Continue(ast.Node, text.Reader, parser.Context) parser.State {
	return parser.Close
}

// Close implements parser.BlockParser.
func (*detailsParser) Close(ast.Node, text.Reader, parser.Context) {}

// CanInterruptParagraph: true, like gno-frame.
func (*detailsParser) CanInterruptParagraph() bool { return true }

// CanAcceptIndentedLine: false; 4+ spaces is an indented code block.
func (*detailsParser) CanAcceptIndentedLine() bool { return false }

// ----- AST transformer -----

type detailsASTTransformer struct{}

// Transform moves the blocks after each open marker under it, up to its
// close marker (removed), a gno-columns or gno-frame marker, or the end of
// its parent; and gives back the depth of the blocks left open at EOF. It
// runs after the frame transformer, so the markers have their final parent.
func (*detailsASTTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	if pc.Get(detailsStackKey) == nil {
		return // no details opener on this page
	}
	endDetails(pc, detailsStack(pc), 0)
	wrapDetails(doc)
}

// wrapDetails gives each open marker under n its blocks, innermost first:
// the walk reaches an inner block's markers through its parent's children
// before its parent's turn, and a parent then takes the inner block whole.
func wrapDetails(n ast.Node) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		wrapDetails(c)
	}
	for c := n.LastChild(); c != nil; c = c.PreviousSibling() {
		if d, ok := c.(*GnoDetailsNode); ok && d.tag == detailsTagOpen {
			takeBlocks(n, d)
		}
	}
}

// takeBlocks moves d's following siblings under d up to its close marker,
// which it removes, or up to a grid or frame marker, which a details block
// must not split, leaving the close marker to render as a stray one. A
// tight list renders its paragraphs as bare text blocks; one moved under d
// becomes a paragraph again.
func takeBlocks(parent ast.Node, d *GnoDetailsNode) {
	for c := d.NextSibling(); c != nil; c = d.NextSibling() {
		if c == d.close {
			parent.RemoveChild(parent, c)
			return
		}
		if k := c.Kind(); k == KindGnoColumn || k == KindGnoFrame {
			break
		}
		if tb, ok := c.(*ast.TextBlock); ok {
			p := ast.NewParagraph()
			p.SetLines(tb.Lines())
			for ch := tb.FirstChild(); ch != nil; ch = tb.FirstChild() {
				p.AppendChild(p, ch)
			}
			parent.RemoveChild(parent, tb)
			d.AppendChild(d, p)
			continue
		}
		d.AppendChild(d, c)
	}
	if d.close != nil {
		d.close.tag = detailsTagInvalid
	}
}

// ----- renderer -----

type detailsRendererHTML struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (*detailsRendererHTML) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGnoDetails, renderGnoDetails)
}

// renderGnoDetails writes constant markup and the escaped summary: no other
// source byte reaches the wrapper. The summary sits in a <span> so the
// summary row's flex layout treats it as one item beside the chevron.
func renderGnoDetails(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*GnoDetailsNode)
	switch {
	case n.tag != detailsTagOpen:
		if entering {
			w.WriteString("<!-- unexpected/invalid details tag omitted -->\n")
		}
	case entering:
		w.WriteString("<details class=\"gno-details\"")
		if n.Open {
			w.WriteString(" open")
		}
		w.WriteString(">\n<summary><svg><use href=\"#ico-arrow\"></use></svg><span>")
		if len(n.Summary) == 0 {
			w.WriteString("Details")
		} else {
			w.Write(util.EscapeHTML(n.Summary))
		}
		w.WriteString("</span></summary>\n<div>\n")
	default:
		w.WriteString("</div>\n</details>\n")
	}
	return ast.WalkContinue, nil
}

// ----- extension registration -----

type details struct{}

// ExtDetails is the singleton Extender for `<gno-details>`.
var ExtDetails = &details{}

// Extend registers the details parser, transformer and renderer.
func (*details) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&detailsParser{}, 498),
		),
		parser.WithASTTransformers(
			// After the frame transformer (501), which moves markers
			// into a frame.
			util.Prioritized(&detailsASTTransformer{}, 502),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&detailsRendererHTML{}, 500),
	))
}
