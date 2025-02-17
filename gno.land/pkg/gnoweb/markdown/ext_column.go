package markdown

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Define NodeKind for Column
var KindColumn = ast.NewNodeKind("Column")

type ColumnTag int

const (
	ColumnTagUndefined ColumnTag = iota
	ColumnTagOpen
	ColumnTagSep
	ColumnTagClose
)

var columnTagNames = map[ColumnTag]string{
	ColumnTagUndefined: "ColumnTagUndefined",
	ColumnTagOpen:      "ColumnTagOpen",
	ColumnTagSep:       "ColumnTagSep",
	ColumnTagClose:     "ColumnTagClose",
}

// ColumnNode represents a semantic tree for "column".
type ColumnNode struct {
	ast.BaseBlock
	Index int
	Tag   ColumnTag
	ctx   *columnContext
}

// Len function returns the length of the context index if it exists
func (cn *ColumnNode) Len() int {
	if cn.ctx != nil {
		return cn.ctx.index
	}
	return 0
}

// Dump implements Node.Dump.
func (n *ColumnNode) Dump(source []byte, level int) {
	kv := make(map[string]string)
	kv["tag"] = columnTagNames[n.Tag]
	kv["head_ref"] = strconv.Itoa(n.ctx.headingLevel)
	if n.Tag == ColumnTagSep {
		kv["index"] = strconv.Itoa(n.Index)
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

// Kind implements Node.Kind.
func (n *ColumnNode) Kind() ast.NodeKind {
	return KindColumn
}

// NewColumn initializes a ColumnAST object.
func NewColumn(ctx *columnContext, tag ColumnTag) *ColumnNode {
	node := &ColumnNode{ctx: ctx, Index: 1, Tag: tag}
	if ctx != nil {
		node.Index = ctx.index
	}
	return node
}

// columnParser struct and its methods are used for parsing columns.
type columnParser struct{}

// Trigger returns the trigger characters for the parser.
func (s *columnParser) Trigger() []byte {
	return []byte{':', '<', '#'}
}

var columnContextKey = parser.NewContextKey()

// columnContext struct and its methods are used for handling column context.
type columnContext struct {
	initilized   bool
	index        int
	headingLevel int // serve as level reference for separator
}

func (ctx *columnContext) Init() {
	ctx.initilized = true
}

func (ctx *columnContext) Destroy() {
	ctx.initilized = false
}

func (ctx *columnContext) SpanColumn() { ctx.index++ }

// parseSeparator checks if the line starts with any of the given separators and returns the separator kind.
func parseSeparator(ctx *columnContext, line []byte) ColumnTag {
	line = util.TrimRightSpace(util.TrimLeftSpace(line))
	if len(line) == 0 {
		return ColumnTagUndefined
	}

	switch string(line) {
	case "<gno-columns>":
		return ColumnTagOpen
	case "</gno-columns>":
		return ColumnTagClose
	case ":::":
		if ctx.initilized {
			return ColumnTagClose
		}

		return ColumnTagOpen
	}

	if line[0] == '#' {
		return ColumnTagSep
	}

	return ColumnTagUndefined
}

func getColumnContext(pc parser.Context) *columnContext {
	cctx, ok := pc.Get(columnContextKey).(*columnContext)
	if !ok || !cctx.initilized {
		cctx = &columnContext{}
		pc.Set(columnContextKey, cctx)
	}

	return cctx
}

// Open function opens a new column node based on the separator kind.
func (s *columnParser) Open(self ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	cctx := getColumnContext(pc)

	line, segment := reader.PeekLine()

	tagKind := parseSeparator(cctx, line)
	if tagKind == ColumnTagUndefined {
		return nil, parser.NoChildren
	}

	node := NewColumn(cctx, tagKind)
	switch tagKind {
	case ColumnTagSep:
		if !cctx.initilized {
			return nil, parser.NoChildren
		}

		level, maxLevel := 0, min(len(line), 6) // maximum level heading is 6
		for level < maxLevel && line[level] == '#' {
			level++
		}

		if cctx.headingLevel == 0 {
			// Register first header as reference
			cctx.headingLevel = level
		} else if cctx.headingLevel != level {
			// If heading level reference is different, skip it
			return nil, parser.NoChildren
		}

		cctx.SpanColumn()

		if trimed := util.TrimLeft(line[level:], []byte{' ', '\n'}); len(trimed) == 0 {
			// empty heading, create a column separtor and skip the parsing
			reader.Advance(segment.Len())
			return node, parser.NoChildren
		}

		// Insert a column separtor but return an empty node so we can
		// let the parser parse the heading
		self.InsertBefore(self, self.PreviousSibling(), node)
		return nil, parser.NoChildren

	case ColumnTagOpen:
		if cctx.initilized {
			return nil, parser.NoChildren
		}

		cctx.Init()

	case ColumnTagClose:
		if !cctx.initilized {
			return nil, parser.NoChildren
		}

		cctx.Destroy()

		reader.Advance(segment.Len())
	}

	return node, parser.NoChildren
}

func (b *columnParser) Continue(n ast.Node, reader text.Reader, _ parser.Context) parser.State {
	return parser.Close
}

func (b *columnParser) Close(_ ast.Node, reader text.Reader, _ parser.Context) {

}

func (b *columnParser) CanInterruptParagraph() bool {
	return true
}

func (b *columnParser) CanAcceptIndentedLine() bool {
	return false
}

func (s *columnParser) CloseBlock(_ ast.Node, _ parser.Context) {}

// columnRender function is used to render the column node.
func columnRender(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	cnode, ok := node.(*ColumnNode)
	numColumns := cnode.Len()
	if !ok || numColumns == 0 || !entering {
		return ast.WalkContinue, nil
	}

	switch cnode.Tag {
	case ColumnTagOpen:
		var classes = []string{
			"gno-cols",
		}

		fmt.Fprintf(w, `<div class="%s">`+"\n", strings.Join(classes, " "))
	case ColumnTagSep:
		if cnode.Index > 0 {
			fmt.Fprint(w, "</div>\n")
		}

		var classes = []string{
			"gno-col",
		}

		fmt.Fprintf(w, "<!-- Column %d -->\n", cnode.Index+1)
		fmt.Fprintf(w, `<div class="%s">`+"\n", strings.Join(classes, " "))
	case ColumnTagClose:
		fmt.Fprint(w, "</div>\n")
		fmt.Fprint(w, "</div>\n")
	default:
		panic("invalid column tag - should not happend")
	}

	return ast.WalkContinue, nil
}

type columnASTTransformer struct{}

func (a *columnASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	cctx := getColumnContext(pc)
	if !cctx.initilized {
		return
	}

	// node hasn't been closed
	lc := node.LastChild()

	nodeCol := NewColumn(cctx, ColumnTagClose)
	cctx.Destroy()
	lc.InsertAfter(lc, lc, nodeCol)

}

type columnRenderer struct{}

// RegisterFuncs add AST objects to Renderer.
func (r *columnRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindColumn, columnRender)
}

// column struct is used to extend the markdown with column functionality.
type column struct{}

var Column = &column{}

// Extend function extends the markdown with column functionality.
func (e *column) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&columnParser{}, 500),
		),
		parser.WithASTTransformers(
			util.Prioritized(&columnASTTransformer{}, 500),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&columnRenderer{}, 500),
	))

}
