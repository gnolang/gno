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

var (
	ErrInvalidColumnsFormat = errors.New("invalid columns format")
	KindColumn              = ast.NewNodeKind("Column")
	MaxHeading              = 6 // Max heading level for markdown
)

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

// ColumnNode represents a semantic tree for a "column".
type ColumnNode struct {
	ast.BaseBlock
	Index  int       // Index of the current column; 0 indicates no column.
	Tag    ColumnTag // Current Column Tag for this node.
	Ignore bool      // Ignore this node, generally meaning it is invalid.

	ctx *columnsContext
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
	if err := n.Error(); err != nil {
		kv["error"] = err.Error()
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

// Kind implements Node.Kind.
func (*ColumnNode) Kind() ast.NodeKind {
	return KindColumn
}

func (c *ColumnNode) String() string {
	return columnTagNames[c.Tag]
}

// Error returns a non-empty error if any error was encountered during parsing.
func (n *ColumnNode) Error() error {
	if n.ctx != nil {
		return n.ctx.Error
	}
	return nil
}

// NewColumn initializes a ColumnNode object.
func NewColumn(ctx *columnsContext, index int, tag ColumnTag) *ColumnNode {
	return &ColumnNode{ctx: ctx, Index: index, Tag: tag}
}

var columnContextKey = parser.NewContextKey()

// columnsContext is used to keep track of columns' state across parsing.
type columnsContext struct {
	IsOpen          bool  // Indicates if a block has been correctly opened.
	Error           error // Error encountered during parsing.
	Index           int   // Current column index.
	RefHeadingLevel int   // Level reference for separators.
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

// columnsParser implements BlockParser.
type columnsParser struct{}

// Trigger returns the trigger characters for the parser.
func (*columnsParser) Trigger() []byte {
	return []byte{'<', '#'}
}

// Open creates a column node based on the line tag.
// If it matches a column tag, it integrates the node into the AST.
func (p *columnsParser) Open(self ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	// Columns tag cannot be a child of another node
	if self.Parent() != nil {
		return nil, parser.NoChildren
	}

	// Get column context
	columnContext, ok := pc.Get(columnContextKey).(*columnsContext)
	if !ok || !columnContext.IsOpen || columnContext.Error != nil {
		columnContext = &columnsContext{} // new context
		pc.Set(columnContextKey, columnContext)
	}

	line, segment := reader.PeekLine()
	tagKind := parseLineTag(line)
	if tagKind == ColumnTagUndefined {
		return nil, parser.NoChildren
	}

	node := NewColumn(columnContext, columnContext.Index, tagKind)
	switch tagKind {
	case ColumnTagSep:
		if !columnContext.IsOpen {
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
		case columnContext.RefHeadingLevel == 0:
			// Register first header as reference
			columnContext.RefHeadingLevel = level
		case columnContext.RefHeadingLevel != level:
			// If heading level reference is different, skip it
			return nil, parser.NoChildren
		}

		// Process creating a column
		columnContext.Index++
		node.Index = columnContext.Index

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
		if columnContext.IsOpen {
			node.Ignore = true
			// Block already open
			return node, parser.NoChildren
		}

		columnContext.IsOpen = true

	case ColumnTagClose:
		if !columnContext.IsOpen {
			node.Ignore = true
			// Block closing without being open
			return node, parser.NoChildren
		}

		columnContext.IsOpen = false
	}

	return node, parser.NoChildren
}

func (*columnsParser) Continue(n ast.Node, reader text.Reader, _ parser.Context) parser.State {
	return parser.Close
}

func (*columnsParser) Close(_ ast.Node, reader text.Reader, _ parser.Context) {}

// CanInterruptParagraph should return true if the parser can interrupt paragraphs.
func (*columnsParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine should return true if the parser can open new nodes when
// the given line is indented more than 3 spaces.
func (*columnsParser) CanAcceptIndentedLine() bool {
	return false
}

// columnRenderer implements NodeRenderer.
// See https://pkg.go.dev/github.com/yuin/goldmark/renderer#NodeRenderer
var _ renderer.NodeRenderer = (*columnsRenderer)(nil)

type columnsRenderer struct{}

// RegisterFuncs adds AST objects to Renderer.
func (r *columnsRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindColumn, columnsRender)
}

// columnsRender function is used to render the column node.
func columnsRender(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	cnode, ok := node.(*ColumnNode)
	if !ok || !entering {
		return ast.WalkContinue, nil
	}

	ignore := cnode.Ignore
	if err := cnode.Error(); err != nil {
		if cnode.Tag == ColumnTagOpen { // only display error on the first tag
			fmt.Fprintf(w, "<!-- gno-columns error: %s -->\n", err.Error())
		}

		ignore = true
	}

	if ignore {
		fmt.Fprintf(w, "<!-- unexpected/invalid %q omitted -->\n", cnode.String())
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
var _ parser.ASTTransformer = (*columnsASTTransformer)(nil)

type columnsASTTransformer struct{}

func (a *columnsASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
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
		if col.ctx.IsOpen {
			col.ctx.Error = fmt.Errorf(
				"%w: columns hasn't been closed", ErrInvalidColumnsFormat,
			)

			continue
		}

		// Check if the first separator is followed by any tag
		if next := n.NextSibling(); next.Kind() != KindColumn {
			col.ctx.Error = fmt.Errorf(
				"%w: open tag should be followed by heading separator or a closing tag",
				ErrInvalidColumnsFormat,
			)
		}
	}
}

type columns struct{}

// column implements Extender
var _ goldmark.Extender = (*columns)(nil)

var Columns = &columns{}

// Extend adds column functionality to the markdown processor.
// XXX: Use 500 for priority for now; we will rework these numbers once another extension is implemented.
func (e *columns) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&columnsParser{}, 500),
		),
		parser.WithASTTransformers(
			util.Prioritized(&columnsASTTransformer{}, 500),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&columnsRenderer{}, 500),
	))
}
