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

// ErrInvalidColumnFormat indicates an invalid columns format.
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
	ctx   *columnContext // context between blocks is unique.
}

// Dump implements Node.Dump for debug representation.
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

// columnParser implement BlockParser.
// See https://pkg.go.dev/github.com/yuin/goldmark/parser#BlockParser
var _ parser.BlockParser = (*columnParser)(nil)

// columnParser struct and its methods are used for parsing columns.
type columnParser struct{}

// Trigger returns the trigger characters for the parser.
func (*columnParser) Trigger() []byte {
	return []byte{':', '<', '#'}
}

var columnContextKey = parser.NewContextKey()

// columnContext struct and its methods are used for handling column context.
type columnContext struct {
	IsOpen          bool
	Error           error
	Index           int
	RefHeadingLevel int // serves as a level reference for separators
}

// parseLineTag checks if the line starts with any of the tag.
func parseLineTag(line []byte) ColumnTag {
	const MaxHeading = 6

	line = util.TrimRightSpace(util.TrimLeftSpace(line))

	// Check if line is a valid heading to treat it as a separator
	if len(line) > 0 && line[0] == '#' && len(line) <= MaxHeading {
		return ColumnTagSep
	}

	switch string(line) {
	case "<gno-columns>":
		return ColumnTagOpen
	case "</gno-columns>":
		return ColumnTagClose
	}

	return ColumnTagUndefined
}

// Open create a column node based on line tag.
func (p *columnParser) Open(self ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	// Columns tag cannot be a child of another node
	if self.Parent() != nil {
		return nil, parser.NoChildren
	}

	// Get column context
	cctx, ok := pc.Get(columnContextKey).(*columnContext)
	if !ok || !cctx.IsOpen {
		cctx = &columnContext{} // new context
		pc.Set(columnContextKey, cctx)
	}

	line, segment := reader.PeekLine()

	tagKind := parseLineTag(line)
	if tagKind == ColumnTagUndefined {
		return nil, parser.NoChildren
	}

	node := NewColumn(cctx, cctx.Index, tagKind)
	switch tagKind {
	case ColumnTagSep:
		if !cctx.IsOpen {
			return nil, parser.NoChildren
		}

		level := 1
		for level < len(line) && line[level] == '#' {
			level++
		}

		switch {
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
		if cctx.IsOpen {
			cctx.IsOpen = false
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

// CanInterruptParagraph should return true if the parser can interrupt paragraphs.
func (*columnParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine should return true if the parser can open new node when
// the given line is being indented more than 3 spaces.
func (*columnParser) CanAcceptIndentedLine() bool {
	return false
}

// columnRenderer implement NodeRenderer
// see https://pkg.go.dev/github.com/yuin/goldmark/renderer#NodeRenderer
var _ renderer.NodeRenderer = (*columnRenderer)(nil)

type columnRenderer struct{}

// RegisterFuncs adds AST objects to Renderer.
func (r *columnRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindColumn, columnRender)
}

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
		if cnode.Index > 0 { // at last one separator
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintln(w, "</div>")

	default:
		panic("invalid column tag - should not happen")
	}

	return ast.WalkContinue, nil
}

// columnASTTransformer implement ASTTransformer.
// See https://pkg.go.dev/github.com/yuin/goldmark/parser#ASTTransformer
var _ parser.ASTTransformer = (*columnASTTransformer)(nil)

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

		// Check if columns block is correctly closed
		if col.ctx.IsOpen {
			col.ctx.Error = fmt.Errorf(
				"%w: columns hasn't been closed", ErrInvalidColumnFormat,
			)
		}

		// Check if first separator is followed by any tag
		if next := n.NextSibling(); next.Kind() != KindColumn {
			col.ctx.Error = fmt.Errorf(
				"%w: open tag should be followed by heading separator", ErrInvalidColumnFormat,
			)
		}
	}
}

// column struct is used to extend the markdown with column functionality.
type column struct{}

// GnoExtension is a goldmark Extender
var _ goldmark.Extender = (*column)(nil)

// Column is an instance of column for extending markdown.
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
