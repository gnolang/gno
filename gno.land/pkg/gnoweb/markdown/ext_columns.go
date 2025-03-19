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

var KindColumn = ast.NewNodeKind("Column")

// ColumnTag represents the type of tag in a column block.
type ColumnTag int

const (
	ColumnTagUndefined ColumnTag = iota // Undefined column tag
	ColumnTagOpen                       // Opening tag for columns
	ColumnTagSep                        // Separator tag for columns
	ColumnTagClose                      // Closing tag for columns
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
	Index int       // Index represents the current column index; 0 indicates no column.
	Tag   ColumnTag // Tag represents the current Column Tag for this node.

	ctx *columnContext
}

// Dump implements Node.Dump for debug representation.
func (n *ColumnNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"tag":      columnTagNames[n.Tag],
		"head_ref": strconv.Itoa(n.ctx.refHeadingLevel),
	}
	if n.Tag == ColumnTagSep {
		kv["index"] = strconv.Itoa(n.Index)
	}
	if err := n.Error(); err != nil {
		kv["error"] = err.Error()
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

// Kind implements Node.Kind.
func (*ColumnNode) Kind() ast.NodeKind {
	return KindColumn
}

// Error returns a non-empty error if any error was encountered during parsing.
func (n *ColumnNode) Error() error {
	if n.ctx != nil {
		return n.ctx.error
	}
	return nil
}

// NewColumn initializes a ColumnNode object.
func NewColumn(ctx *columnContext, index int, tag ColumnTag) *ColumnNode {
	return &ColumnNode{ctx: ctx, Index: index, Tag: tag}
}

var columnContextKey = parser.NewContextKey()

// columnContext is used to keep track of columns' state across parsing.
type columnContext struct {
	isOpen          bool  // Indicates if a block has been correctly opened
	error           error // Error encountered during parsing
	index           int   // Current column index
	refHeadingLevel int   // Level reference for separators
}

// parseLineTag checks if the line matches open or closing tag or if the line starts with a heading.
func parseLineTag(line []byte) ColumnTag {
	line = util.TrimRightSpace(util.TrimLeftSpace(line))

	if len(line) > 0 && line[0] == '#' {
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

// columnParser implements BlockParser.
type columnParser struct{}

// Trigger returns the trigger characters for the parser.
func (*columnParser) Trigger() []byte {
	return []byte{'<', '#'}
}

// Open creates a column node based on the line tag.
// If it matches a column tag, it integrates the node into the AST.
func (p *columnParser) Open(self ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	const MaxHeading = 6

	// Columns tag cannot be a child of another node
	if self.Parent() != nil {
		return nil, parser.NoChildren
	}

	// Get column context
	cctx, ok := pc.Get(columnContextKey).(*columnContext)
	if !ok || !cctx.isOpen || cctx.error != nil {
		cctx = &columnContext{} // new context
		pc.Set(columnContextKey, cctx)
	}

	line, segment := reader.PeekLine()
	tagKind := parseLineTag(line)
	if tagKind == ColumnTagUndefined {
		return nil, parser.NoChildren
	}

	node := NewColumn(cctx, cctx.index, tagKind)
	switch tagKind {
	case ColumnTagSep:
		if !cctx.isOpen {
			return nil, parser.NoChildren
		}

		level, maxLevel := 1, min(len(line), MaxHeading+1)
		for level < maxLevel && line[level] == '#' {
			level++
		}

		switch {
		case level > MaxHeading:
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
		cctx.index++
		node.Index = cctx.index

		// Check for non-empty heading
		if trimmed := util.TrimLeftSpace(line[level:]); len(trimmed) > 0 {
			// Insert a column separator but return an empty node so we can
			// let the parser parse the heading
			self.InsertBefore(self, self.PreviousSibling(), node)
			return nil, parser.NoChildren
		}

		// Empty heading, create a column separator
		reader.Advance(segment.Len())

	case ColumnTagOpen:
		if cctx.isOpen {
			// Block already open
			return nil, parser.NoChildren
		}

		cctx.isOpen = true

	case ColumnTagClose:
		if !cctx.isOpen {
			// Block closing without being open
			return nil, parser.NoChildren
		}

		cctx.isOpen = false
	}

	return node, parser.NoChildren
}

func (*columnParser) Continue(n ast.Node, reader text.Reader, _ parser.Context) parser.State {
	return parser.Close
}

func (*columnParser) Close(_ ast.Node, reader text.Reader, _ parser.Context) {}

// CanInterruptParagraph should return true if the parser can interrupt paragraphs.
func (*columnParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine should return true if the parser can open new nodes when
// the given line is indented more than 3 spaces.
func (*columnParser) CanAcceptIndentedLine() bool {
	return false
}

// columnRenderer implements NodeRenderer
// See https://pkg.go.dev/github.com/yuin/goldmark/renderer#NodeRenderer
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

	if err := cnode.Error(); err != nil {
		if cnode.Tag == ColumnTagOpen { // only display error on the first tag
			fmt.Fprintf(w, "<!-- gno-columns error: %s -->\n", err.Error())
		}
		return ast.WalkContinue, nil
	}

	switch cnode.Tag {
	case ColumnTagOpen:
		fmt.Fprint(w, `<div class="gno-columns">`+"\n")

	case ColumnTagSep:
		if cnode.Index > 1 {
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintf(w, "<!-- Column %d -->\n", cnode.Index)
		fmt.Fprintln(w, "<div>")

	case ColumnTagClose:
		if cnode.Index > 0 { // at least one separator
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintln(w, "</div>")

	default:
		panic("invalid column tag - should not happen")
	}

	return ast.WalkContinue, nil
}

// columnASTTransformer implements ASTTransformer.
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
		if col.Error() != nil || col.Tag != ColumnTagOpen {
			continue
		}

		// Check if columns block is correctly closed
		if col.ctx.isOpen {
			col.ctx.error = fmt.Errorf(
				"%w: columns hasn't been closed", ErrInvalidColumnFormat,
			)

			continue
		}

		// Check if the first separator is followed by any tag
		if next := n.NextSibling(); next.Kind() != KindColumn {
			col.ctx.error = fmt.Errorf(
				"%w: open tag should be followed by heading separator or a closing tag",
				ErrInvalidColumnFormat,
			)
		}
	}
}

type column struct{}

// column implements Extender
var _ goldmark.Extender = (*column)(nil)

var Column = &column{}

// Extend adds column functionality to the markdown processor.
// XXX: Use 500 for priority for now; we will rework these numbers once another extension is implemented.
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
