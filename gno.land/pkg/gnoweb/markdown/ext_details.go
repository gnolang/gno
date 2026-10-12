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
	// refused marks an opener the depth cap turned down; it stays on the
	// parse stack so that its own close tag is the one it takes.
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
// close tag or a self-closing form makes the line invalid rather than
// silently ignored. Text after the tag makes it no tag line at all, so the
// text stays in a paragraph, as with gno-frame. Tag names are
// case-insensitive, like the other gno-* tags.
func parseDetailsLineTag(line []byte) (kind detailsTagKind, open bool, summary []byte) {
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
		return detailsTagNone, false, nil
	case selfClosing || !valid:
		return detailsTagInvalid, false, nil
	}
	return kind, seenOpen, summary
}

// ----- parse state -----

// detailsStackKey holds the openers not yet closed, refused ones included,
// innermost last.
var detailsStackKey = parser.NewContextKey()

func detailsStack(pc parser.Context) []*GnoDetailsNode {
	stack, _ := pc.Get(detailsStackKey).([]*GnoDetailsNode)
	return stack
}

// ----- block parser -----

type detailsParser struct{}

var _ parser.BlockParser = (*detailsParser)(nil)

func (*detailsParser) Trigger() []byte { return []byte{'<'} }

// Open reads a details tag at any level, so a block may sit in a list item,
// a blockquote, an alert or another details block. An opener takes a level
// of the gno-* depth cap and is refused past it. A close tag pairs with the
// innermost opener of the same parent and gives its level back, along with
// every opener after it, whose own close never came. A refused opener, its
// close, a stray close tag and a malformed tag each render as a comment.
// Every tag line yields a leaf, never nil: nil would hand the line to the
// type-7 HTML block parser, which swallows every line up to the next blank
// one.
func (*detailsParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	// Goldmark never offers a line indented 4 columns or more; a tab after a
	// container prefix counts to its next stop.
	_, pos := util.IndentWidth(line, reader.LineOffset())
	kind, open, summary := parseDetailsLineTag(util.TrimRightSpace(line[pos:]))
	if kind == detailsTagNone {
		return nil, parser.NoChildren
	}
	reader.AdvanceToEOL()

	node := &GnoDetailsNode{tag: detailsTagInvalid}
	stack := detailsStack(pc)
	switch kind {
	case detailsTagOpen:
		if Push(pc) { // the cap spans all gno-* blocks
			node.tag, node.Open, node.Summary = detailsTagOpen, open, summary
		} else {
			node.refused = true
		}
		pc.Set(detailsStackKey, append(stack, node))
	case detailsTagClose:
		for i := len(stack) - 1; i >= 0; i-- {
			if stack[i].Parent() != parent {
				continue
			}
			node.tag = detailsTagClose
			stack[i].close = node
			popDetails(pc, stack[i:])
			pc.Set(detailsStackKey, stack[:i])
			break
		}
	}
	return node, parser.NoChildren
}

// popDetails gives back the depth of every accepted opener in openers.
func popDetails(pc parser.Context, openers []*GnoDetailsNode) {
	for _, d := range openers {
		if !d.refused {
			Pop(pc)
		}
	}
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
	popDetails(pc, detailsStack(pc))
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
