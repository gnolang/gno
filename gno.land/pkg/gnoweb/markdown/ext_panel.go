package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

// KindGnoPanel is the node kind of a `<gno-panel>` block.
var KindGnoPanel = ast.NewNodeKind("GnoPanel")

// GnoPanelNode is a framed container of ordinary markdown:
//
//	<gno-panel>
//	## Any markdown
//	</gno-panel>
//
// Standalone it renders as a highlighted panel; alone in a column the
// CSS turns it into a card.
type GnoPanelNode struct {
	ast.BaseBlock
	// invalid marks a stray, malformed or refused tag. It stays a leaf
	// rendered as a comment, so the line never falls through to the
	// type-7 HTML block parser, which would swallow every line up to the
	// next blank one.
	invalid bool
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
	panelOpenTag     = []byte("<gno-panel>")
	panelCloseTag    = []byte("</gno-panel>")
	panelOpenPrefix  = []byte("<gno-panel")
	panelClosePrefix = []byte("</gno-panel")

	// Prefixes of every gno-columns tag, separator included.
	columnsOpenPrefix  = []byte("<gno-columns")
	columnsClosePrefix = []byte("</gno-columns")
)

func hasPrefixFold(s, prefix []byte) bool {
	return len(s) >= len(prefix) && bytes.EqualFold(s[:len(prefix)], prefix)
}

// parsePanelLineTag classifies a trimmed line. Only a bare `<gno-panel>`
// or `</gno-panel>` is valid: the attribute allowlist is empty, so any
// attribute, a self-closing form or trailing text makes the line invalid
// rather than silently ignored. Tag names are case-insensitive, like the
// other gno-* tags (the HTML tokenizer lowercases them).
func parsePanelLineTag(line []byte) panelTagKind {
	// Allocation-free paths first: Continue runs this on every line of a
	// panel, and the HTML tokenizer below costs a few KB per call.
	switch {
	case bytes.EqualFold(line, panelOpenTag):
		return panelTagOpen
	case bytes.EqualFold(line, panelCloseTag):
		return panelTagClose
	case !hasPrefixFold(line, panelOpenPrefix) && !hasPrefixFold(line, panelClosePrefix):
		return panelTagNone
	}
	toks, err := ParseHTMLTokens(bytes.NewReader(line))
	if err != nil || len(toks) == 0 {
		return panelTagNone
	}
	tok := toks[0]
	if tok.Data != "gno-panel" {
		return panelTagNone
	}
	switch {
	case len(toks) != 1 || len(tok.Attr) != 0 || tok.Type == html.SelfClosingTagToken:
		return panelTagInvalid
	case tok.Type == html.StartTagToken:
		return panelTagOpen
	default:
		return panelTagClose
	}
}

// ----- block parser -----

type panelParser struct{}

var _ parser.BlockParser = (*panelParser)(nil)

func (*panelParser) Trigger() []byte { return []byte{'<'} }

// Open opens a panel on `<gno-panel>`. A close tag here is stray (an open
// panel consumes its own close in Continue); it, a malformed tag, a panel
// inside a panel and an opener past the nesting cap all yield an invalid
// leaf.
func (*panelParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	kind := parsePanelLineTag(trimTagLine(line))
	if kind == panelTagNone {
		return nil, parser.NoChildren
	}

	reader.AdvanceToEOL()
	// Push last: it must only run for a panel that opens. The nesting cap
	// spans all structural Gno blocks.
	if kind != panelTagOpen || hasPanelAncestor(parent) || !Push(pc) {
		return &GnoPanelNode{invalid: true}, parser.NoChildren
	}
	return &GnoPanelNode{}, parser.HasChildren
}

func hasPanelAncestor(n ast.Node) bool {
	for ; n != nil; n = n.Parent() {
		if n.Kind() == KindGnoPanel {
			return true
		}
	}
	return false
}

// Continue closes the panel on `</gno-panel>`, consuming the line, and
// on any gno-columns tag, leaving the line to reopen at the parent so an
// unclosed panel cannot swallow its column's separator.
//
// goldmark asks the panel before its children, so lines owned by an open
// opaque child must pass through untouched: a fenced code block may show
// the syntax, and a <gno-foreign> body is untrusted bytes that must not
// close the host's panel and spill out of the sandbox.
func (*panelParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	if node.(*GnoPanelNode).invalid {
		return parser.Close
	}
	switch pc.LastOpenedBlock().Node.Kind() {
	case ast.KindFencedCodeBlock, KindGnoForeign:
		return parser.Continue | parser.HasChildren
	}

	line, _ := reader.PeekLine()
	line = trimTagLine(line)
	if len(line) == 0 || line[0] != '<' {
		return parser.Continue | parser.HasChildren
	}
	if parsePanelLineTag(line) == panelTagClose {
		reader.AdvanceToEOL()
		return parser.Close
	}
	if (hasPrefixFold(line, columnsOpenPrefix) || hasPrefixFold(line, columnsClosePrefix)) &&
		parseLineTag(line) != GnoColumnTagUndefined {
		return parser.Close
	}
	return parser.Continue | parser.HasChildren
}

// Close pops the depth pushed by Open. goldmark calls Close once per
// opened block, EOF included, so an unclosed panel stays balanced.
func (*panelParser) Close(node ast.Node, _ text.Reader, pc parser.Context) {
	if !node.(*GnoPanelNode).invalid {
		Pop(pc)
	}
}

// CanInterruptParagraph: true, like gno-columns.
func (*panelParser) CanInterruptParagraph() bool { return true }

// CanAcceptIndentedLine: false; 4+ spaces is an indented code block.
func (*panelParser) CanAcceptIndentedLine() bool { return false }

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
	case node.(*GnoPanelNode).invalid:
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

// Extend registers the panel parser and renderer.
func (*panels) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(&panelParser{}, 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&panelRendererHTML{}, 500),
	))
}
