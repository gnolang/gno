package markdown

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var ErrInvalidColumnFormat = errors.New("invalid columns format")

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

	// Context bewteen block is uniq.
	// A block is composed of an opening, some potential separator and a
	// closing.
	ctx *columnContext
}

// Dump implements Node.Dump.
func (n *ColumnNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"tag":      columnTagNames[n.Tag],
		"head_ref": strconv.Itoa(n.ctx.RefHeadingLevel),
	}
	if n.Tag == ColumnTagSep {
		kv["index"] = strconv.Itoa(n.Index)
	}
	if n.ctx.Error != nil {
		kv["error"] = n.ctx.Error.Error()
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
	IsClose, IsOpen bool
	Error           error
	Index           int
	RefHeadingLevel int // serves as a level reference for separators
}

func isPreviousNodeTag(node ast.Node, tag ColumnTag) bool {
	if node == nil {
		return false
	}

	if cnode, ok := node.(*ColumnNode); ok {
		return cnode.Tag == tag
	}

	return false
}

func getColumnContext(pc parser.Context) *columnContext {
	cctx, ok := pc.Get(columnContextKey).(*columnContext)
	if !ok || cctx.IsClose || cctx.Error != nil {
		cctx = &columnContext{} // new context
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
		if ctx.IsOpen {
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

	// Columns tag cannot be a children
	if self.Parent() != nil {
		return nil, parser.NoChildren
	}

	cctx := getColumnContext(pc)
	if cctx.Error != nil {
		// Don't bother with malformed block
		return nil, parser.NoChildren
	}

	line, segment := reader.PeekLine()

	tagKind := parseLineTag(cctx, line)
	if tagKind == ColumnTagUndefined {
		return nil, parser.NoChildren
	}

	node := NewColumn(cctx, cctx.Index, tagKind)
	switch tagKind {
	case ColumnTagSep:
		if !cctx.IsOpen {
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
		case cctx.RefHeadingLevel == 0:
			// Register first header as reference
			cctx.RefHeadingLevel = level
		case cctx.RefHeadingLevel != level:
			// If heading level reference is different, skip it
			return nil, parser.NoChildren
		}

		// Process creating a column
		cctx.Index++
		node.Index = cctx.Index

		if trimmed := util.TrimLeft(line[level:], []byte{' ', '\n'}); len(trimmed) == 0 {
			// Empty heading, create a column separator and skip the parsing
			reader.Advance(segment.Len())
			return node, parser.NoChildren
		}

		// Insert a column separator but return an empty node so we can
		// let the parser parse the heading
		self.InsertBefore(self, self.PreviousSibling(), node)

	case ColumnTagOpen:
		if !cctx.IsOpen {
			cctx.IsOpen = true
			return node, parser.NoChildren
		}

	case ColumnTagClose:
		if cctx.IsOpen && !cctx.IsClose {
			cctx.IsClose = true
			return node, parser.NoChildren
		}
	}

	// Ignore node
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

	if err := cnode.ctx.Error; err != nil {
		if cnode.Tag == ColumnTagOpen {
			fmt.Fprintf(w, "<!-- gno-columns error: %s -->\n", err.Error())
		}

		return ast.WalkContinue, nil
	}

	switch cnode.Tag {
	case ColumnTagOpen:
		fmt.Fprint(w, `<div class="gno-cols">`+"\n")

	case ColumnTagSep:
		if cnode.Index > 1 {
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintf(w, "<!-- Column %d -->\n", cnode.Index)
		fmt.Fprintln(w, "<div>")

	case ColumnTagClose:
		prev, ok := cnode.PreviousSibling().(*ColumnNode)
		if !ok || prev.Tag != ColumnTagOpen {
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintln(w, "</div>")

	default:
		panic("invalid column tag - should not happen")
	}

	return ast.WalkContinue, nil
}

type columnASTTransformer struct{}

func (a *columnASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	// Validate columns
	for n := node.FirstChild(); n != nil; n = n.NextSibling() {
		if n.Kind() != KindColumn {
			continue
		}

		col := n.(*ColumnNode)
		if col.Tag != ColumnTagOpen {
			continue
		}

		// Check if columns block is correctly close
		if !col.ctx.IsClose {
			col.ctx.Error = fmt.Errorf(
				"%w: columns hasn't been closed", ErrInvalidColumnFormat,
			)

		}

		// Check if first sperator is followed by any tag
		if next := n.NextSibling(); next.Kind() != KindColumn {
			col.ctx.Error = fmt.Errorf(
				"%w: open tag should be followed by heading separtor", ErrInvalidColumnFormat,
			)
		}
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
