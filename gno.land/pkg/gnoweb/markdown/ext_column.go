package markdown

import (
	"fmt"
	"strconv"

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

// Dump implements Node.Dump.
func (n *ColumnNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"tag":      columnTagNames[n.Tag],
		"head_ref": strconv.Itoa(n.ctx.refHeadingLevel),
		"depth":    strconv.Itoa(n.ctx.depth),
	}
	if n.Tag == ColumnTagSep {
		kv["index"] = strconv.Itoa(n.Index)
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

// Kind implements Node.Kind.
func (*ColumnNode) Kind() ast.NodeKind {
	return KindColumn
}

// NewColumn initializes a ColumnNode object.
func NewColumn(ctx *columnContext, index int, tag ColumnTag) *ColumnNode {
	return &ColumnNode{ctx: ctx, Index: index, Tag: tag}
}

// columnParser struct and its methods are used for parsing columns.
type columnParser struct{}

// Trigger returns the trigger characters for the parser.
func (*columnParser) Trigger() []byte {
	return []byte{':', '<', '#'}
}

var columnContextKey = parser.NewContextKey()

// columnContext struct and its methods are used for handling column context.
type columnContext struct {
	initialized     bool
	prevContext     *columnContext
	openNode        ast.Node
	index           int
	refHeadingLevel int // serves as a level reference for separators
	depth           int
}

func (ctx *columnContext) Init(node ast.Node) (succeed bool) {
	ctx.depth++
	if ctx.depth == 1 {
		ctx.openNode = node
		return true
	}

	return false
}

func (ctx *columnContext) IsInitilized() bool {
	return ctx.depth >= 1
}

func (ctx *columnContext) Destroy() (succeed bool) {
	ctx.depth--
	return ctx.depth == 0
}

func (ctx *columnContext) SpanColumn() {
	ctx.index++
}

func getColumnContext(pc parser.Context) *columnContext {
	cctx, ok := pc.Get(columnContextKey).(*columnContext)
	switch {
	case !ok:
		cctx = &columnContext{index: 1}
	case !cctx.IsInitilized():
		cctx = &columnContext{prevContext: cctx, index: 1}
	default:
		return cctx
	}

	pc.Set(columnContextKey, cctx)
	return cctx
}

// parseLineTag checks if the line starts with any of the tag.
func parseLineTag(ctx *columnContext, line []byte) ColumnTag {
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
		if ctx.IsInitilized() {
			return ColumnTagClose
		}
		return ColumnTagOpen
	}

	if line[0] == '#' {
		return ColumnTagSep
	}

	return ColumnTagUndefined
}

// Open opens a new column node based on the separator kind.
func (p *columnParser) Open(self ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	const MaxHeading = 6
	if self.Parent() != nil {
		return nil, parser.NoChildren
	}

	cctx := getColumnContext(pc)
	line, segment := reader.PeekLine()

	tagKind := parseLineTag(cctx, line)
	if tagKind == ColumnTagUndefined {
		return nil, parser.NoChildren
	}

	node := NewColumn(cctx, cctx.index, tagKind)
	switch tagKind {
	case ColumnTagSep:
		if !cctx.IsInitilized() {
			return nil, parser.NoChildren
		}

		level, maxLevel := 1, min(len(line), MaxHeading+1)
		for level < maxLevel && line[level] == '#' {
			level++
		}

		switch {
		case level >= MaxHeading:
			// Level is beyond the maximum one, ignore this heading
			return nil, parser.NoChildren
		case cctx.refHeadingLevel == 0:
			// Register first header as reference
			cctx.refHeadingLevel = level
		case cctx.refHeadingLevel != level:
			// If heading level reference is different, skip it
			return nil, parser.NoChildren
		}

		// Process creating a column
		cctx.SpanColumn()

		if trimmed := util.TrimLeft(line[level:], []byte{' ', '\n'}); len(trimmed) == 0 {
			// Empty heading, create a column separator and skip the parsing
			reader.Advance(segment.Len())
			return node, parser.NoChildren
		}

		// Insert a column separator but return an empty node so we can
		// let the parser parse the heading
		self.InsertBefore(self, self.PreviousSibling(), node)

	case ColumnTagOpen:
		if cctx.Init(node) {
			return node, parser.NoChildren
		}

		reader.Advance(segment.Len())

	case ColumnTagClose:
		if cctx.Destroy() {
			return node, parser.NoChildren
		}

		if cctx.depth > 0 {
			reader.Advance(segment.Len())
		}
	}

	return nil, parser.NoChildren
}

func (*columnParser) Continue(n ast.Node, reader text.Reader, _ parser.Context) parser.State {
	return parser.Close
}

func (*columnParser) Close(_ ast.Node, reader text.Reader, _ parser.Context) {}

func (*columnParser) CanInterruptParagraph() bool {
	return true
}

func (*columnParser) CanAcceptIndentedLine() bool {
	return false
}

func (*columnParser) CloseBlock(_ ast.Node, _ parser.Context) {}

// columnRender function is used to render the column node.
func columnRender(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	cnode, ok := node.(*ColumnNode)
	if !ok || !entering {
		return ast.WalkContinue, nil
	}

	switch cnode.Tag {
	case ColumnTagOpen:
		fmt.Fprint(w, `<div class="gno-cols">`+"\n")

	case ColumnTagSep:
		prev, ok := cnode.PreviousSibling().(*ColumnNode)
		if !ok || prev.Tag != ColumnTagOpen {
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintf(w, "<!-- Column %d -->\n", cnode.Index)
		fmt.Fprintln(w, "<div>")

	case ColumnTagClose:
		fmt.Fprintln(w, "</div>\n</div>")

	default:
		panic("invalid column tag - should not happen")
	}

	return ast.WalkContinue, nil
}

type columnASTTransformer struct{}

func (a *columnASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	cctx, ok := pc.Get(columnContextKey).(*columnContext)
	if !ok {
		return
	}
	defer cctx.Destroy()

	// Check if node hasn't been closed
	if cctx.IsInitilized() {
		// If not closed simply add a closed node at the end
		lc := node.LastChild()

		nodeCol := NewColumn(cctx, cctx.index, ColumnTagClose)
		lc.InsertAfter(lc, lc, nodeCol)
	}

	// Ensure that each open tag always start with a column
	for cctx != nil && cctx.openNode != nil {
		next := cctx.openNode.NextSibling()
		if _, ok := next.(*ColumnNode); !ok {
			// Generate column0
			column0 := NewColumn(cctx, 0, ColumnTagSep)

			// Instert column after OpenNode
			node.InsertAfter(node, cctx.openNode, column0)
		}

		cctx = cctx.prevContext
	}

}

type columnRenderer struct{}

// RegisterFuncs adds AST objects to Renderer.
func (r *columnRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindColumn, columnRender)
}

// column struct is used to extend the markdown with column functionality.
type column struct{}

var Column = &column{}

// Extend extends the markdown with column functionality.
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
