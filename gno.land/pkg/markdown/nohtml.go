package markdown

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Define NodeKind for NoHtml
var KindNoHtml = ast.NewNodeKind("NoHtml")

// Define NoHtmlTag type and constants
type NoHtmlTag int

const (
	NoHtmlTagUndefined NoHtmlTag = iota
	NoHtmlTagOpen
	NoHtmlTagSep
	NoHtmlTagClose
)

// NoHtmlNode represents a semantic tree for "noHtml".
type NoHtmlNode struct {
	ast.BaseBlock
	Index int
	Tag   NoHtmlTag
	ctx   *noHtmlContext
}

// Len function returns the length of the context index if it exists
func (cn *NoHtmlNode) Len() int {
	if cn.ctx != nil {
		return cn.ctx.index
	}
	return 0
}

// Dump implements Node.Dump.
func (n *NoHtmlNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// Kind implements Node.Kind.
func (n *NoHtmlNode) Kind() ast.NodeKind {
	return KindNoHtml
}

// NewNoHtml initializes a NoHtmlAST object.
func NewNoHtml(ctx *noHtmlContext) *NoHtmlNode {
	node := &NoHtmlNode{ctx: ctx, Index: 1}
	if ctx != nil {
		node.Index = ctx.index
	}
	return node
}

// noHtmlParser struct and its methods are used for parsing noHtmls.
type noHtmlParser struct{}

var defaultNoHtmlParser = &noHtmlParser{}

func NewNoHtmlBlockParser() parser.BlockParser {
	return defaultNoHtmlParser
}

// Trigger returns the trigger characters for the parser.
func (s *noHtmlParser) Trigger() []byte {
	return []byte{'<'}
}

var noHtmlContextKey = parser.NewContextKey()

// noHtmlContext struct and its methods are used for handling noHtml context.
type noHtmlContext struct {
	initilized bool
	index      int
}

func (ctx *noHtmlContext) Init() {
	ctx.initilized = true
	ctx.index = 1
}

func (ctx *noHtmlContext) Destroy() { ctx.initilized = false }

func (ctx *noHtmlContext) SpanNoHtml() { ctx.index++ }

func (s *noHtmlParser) getNoHtmlContext(pc parser.Context) *noHtmlContext {
	cctx, ok := pc.Get(noHtmlContextKey).(*noHtmlContext)
	if !ok || !cctx.initilized {
		cctx = &noHtmlContext{}
		pc.Set(noHtmlContextKey, cctx)
	}

	return cctx
}

// Open function opens a new noHtml node based on the separator kind.
func (s *noHtmlParser) Open(_ ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	// line, segment := reader.PeekLine()
	// kind, _, ok := parseSeparator(line)
	// if !ok {
	// 	return nil, parser.NoChildren
	// }

	// cctx := s.getNoHtmlContext(pc)
	// node := NewNoHtml(cctx)
	// switch kind {
	// case '=':
	// 	if !cctx.initilized {
	// 		cctx.Init()
	// 		node.Tag = NoHtmlTagOpen
	// 	} else {
	// 		cctx.Destroy()
	// 		node.Tag = NoHtmlTagClose
	// 	}
	// case '+':
	// 	if !cctx.initilized {
	// 		return nil, parser.HasChildren
	// 	}

	// 	cctx.SpanNoHtml()
	// 	node.Tag = NoHtmlTagSep
	// default:
	// 	panic("invalid tag - should not happen")
	// }

	// reader.Advance(segment.Len())

	// node.Index = cctx.index
	return nil, parser.NoChildren
}

func (b *noHtmlParser) Continue(_ ast.Node, _ text.Reader, _ parser.Context) parser.State {
	return parser.Close
}

func (b *noHtmlParser) Close(_ ast.Node, _ text.Reader, _ parser.Context) {}

func (b *noHtmlParser) CanInterruptParagraph() bool {
	return true
}

func (b *noHtmlParser) CanAcceptIndentedLine() bool {
	return false
}

func (s *noHtmlParser) CloseBlock(_ ast.Node, _ parser.Context) {}

// noHtmlRender function is used to render the noHtml node.
func noHtmlRender(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	cnode, ok := node.(*NoHtmlNode)
	numNoHtmls := cnode.Len()
	if !ok || numNoHtmls == 0 || !entering {
		return ast.WalkContinue, nil
	}

	switch cnode.Tag {
	case NoHtmlTagOpen:
		fmt.Fprintf(w, `<gno-column class="col-%[1]d" col="%[1]d">`+"\n", numNoHtmls)
		fmt.Fprintf(w, "<!-- NoHtml %d -->\n", cnode.Index)
		fmt.Fprint(w, "<div>\n")
	case NoHtmlTagSep:
		fmt.Fprint(w, "<div>\n")
		fmt.Fprintf(w, "<!-- NoHtml %d -->\n", cnode.Index)
		fmt.Fprint(w, "<div>\n")
	case NoHtmlTagClose:
		fmt.Fprint(w, "</div>\n")
		fmt.Fprint(w, "</gno-column>\n")
	default:
		panic("invalid noHtml tag - should not happend")
	}

	return ast.WalkContinue, nil
}

// noHtml struct is used to extend the markdown with noHtml functionality.
type noHtml struct{}

var NoHtml = &noHtml{}

// Extend function extends the markdown with noHtml functionality.
func (e *noHtml) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(NewNoHtmlBlockParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewRenderer(), 500),
	))
}
