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

// Error messages for invalid column tags.
var (
	ErrColumnsInvalidOpenFormat      = errors.New("invalid heading format")
	ErrColumnsUnexpectedOrInvalidTag = errors.New("unexpected or invalid tag")
)

// Define custom node kind and maximum heading level.
var (
	KindGnoColumn = ast.NewNodeKind("GnoColumn")
	MaxHeading    = 6
)

// GnoColumnTag represents the type of tag in a column block.
type GnoColumnTag int

const (
	GnoColumnTagUndefined GnoColumnTag = iota
	GnoColumnTagOpen
	GnoColumnTagSep
	GnoColumnTagClose
)

var GnocolumnTagNames = map[GnoColumnTag]string{
	GnoColumnTagUndefined: "ColumnTagUndefined",
	GnoColumnTagOpen:      "ColumnTagOpen",
	GnoColumnTagSep:       "ColumnTagSep",
	GnoColumnTagClose:     "ColumnTagClose",
}

// GnoColumnNode represents a semantic tree for a "column".
type GnoColumnNode struct {
	ast.BaseBlock
	Index int          // Index of the column associated with the node.
	Tag   GnoColumnTag // Current Column Tag for this node.
	Error error        // If not nil, indicates that the node is invalid.

	ctx *columnsContext
}

// Dump implements Node.Dump for debug representation.
func (n *GnoColumnNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"tag":      GnocolumnTagNames[n.Tag],
		"head_ref": strconv.Itoa(n.ctx.RefHeadingLevel),
	}
	if n.Tag == GnoColumnTagSep {
		kv["index"] = strconv.Itoa(n.Index)
	}
	if err := n.Error; err != nil {
		kv["error"] = err.Error()
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

// Kind implements Node.Kind.
func (*GnoColumnNode) Kind() ast.NodeKind {
	return KindGnoColumn
}

func (n *GnoColumnNode) String() string {
	return GnocolumnTagNames[n.Tag]
}

// NewColumn initializes a ColumnNode object.
func NewColumn(ctx *columnsContext, index int, tag GnoColumnTag) *GnoColumnNode {
	return &GnoColumnNode{ctx: ctx, Index: index, Tag: tag}
}

var columnContextKey = parser.NewContextKey()

// columnsContext is used to keep track of columns' state across parsing.
type columnsContext struct {
	IsOpen          bool           // Indicates if a block has been correctly opened.
	Index           int            // Index of the current column; 0 indicates no column.
	RefHeadingLevel int            // Level reference for separators.
	OpenTag         *GnoColumnNode // First opening tag for this context.
}

// parseLineTag identifies the tag type based on the line content.
func parseLineTag(line []byte) GnoColumnTag {
	line = util.TrimRightSpace(util.TrimLeftSpace(line))

	if len(line) > 0 && line[0] == '#' {
		return GnoColumnTagSep
	}

	switch string(line) {
	case "<gno-columns>":
		return GnoColumnTagOpen
	case "</gno-columns>":
		return GnoColumnTagClose
	}

	return GnoColumnTagUndefined
}

// columnsParser implements BlockParser.
var _ parser.BlockParser = (*columnsParser)(nil)

type columnsParser struct{}

// Trigger returns the trigger characters for the parser.
func (*columnsParser) Trigger() []byte {
	return []byte{'<', '#'}
}

// Open creates a column node based on the line tag.
func (p *columnsParser) Open(doc ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	// Columns tag cannot be a child of another node.
	if doc.Parent() != nil {
		return nil, parser.NoChildren
	}

	// Get column context.
	cctx, ok := pc.Get(columnContextKey).(*columnsContext)
	if !ok || !cctx.IsOpen {
		cctx = &columnsContext{} // New context.
		pc.Set(columnContextKey, cctx)
	}

	line, segment := reader.PeekLine()
	tagKind := parseLineTag(line)
	if tagKind == GnoColumnTagUndefined {
		return nil, parser.NoChildren
	}

	node := NewColumn(cctx, cctx.Index, tagKind)
	switch tagKind {
	case GnoColumnTagSep:
		if !cctx.IsOpen {
			return nil, parser.NoChildren
		}

		level, maxLevel := 1, min(len(line), MaxHeading+1)
		for level < maxLevel && line[level] == '#' {
			level++
		}

		switch {
		case level > MaxHeading:
			// Level is beyond the maximum one, ignore this heading.
			return nil, parser.NoChildren
		case cctx.RefHeadingLevel == 0:
			// Register first header as reference.
			cctx.RefHeadingLevel = level
		case cctx.RefHeadingLevel != level:
			// If heading level reference is different, skip it.
			return nil, parser.NoChildren
		}

		// First separator should follow an Open Tag.
		if cctx.Index == 0 {
			lc := doc.LastChild()
			if ln, ok := lc.(*GnoColumnNode); !ok || ln.Tag != GnoColumnTagOpen {
				// Open tag isn't followed by a heading.
				// Mark open tag as wrong and immediately close the context.
				cctx.OpenTag.Error = ErrColumnsInvalidOpenFormat
				cctx.IsOpen = false
				return nil, parser.NoChildren
			}
		}

		// Process creating a column.
		cctx.Index++
		node.Index = cctx.Index

		// Check for non-empty heading.
		if trimmed := util.TrimLeftSpace(line[level:]); len(trimmed) > 0 {
			// Insert a column separator but return an empty node so we can
			// let the parser parse the heading.
			doc.InsertBefore(doc, doc.PreviousSibling(), node)
			return nil, parser.NoChildren
		}

		// Empty heading, create a column separator.
		reader.Advance(segment.Len())

	case GnoColumnTagOpen:
		if cctx.IsOpen {
			// Block already open.
			node.Error = ErrColumnsUnexpectedOrInvalidTag
			return node, parser.NoChildren
		}

		cctx.OpenTag = node
		cctx.IsOpen = true

	case GnoColumnTagClose:
		if !cctx.IsOpen {
			// Block closing without being open.
			node.Error = ErrColumnsUnexpectedOrInvalidTag
			return node, parser.NoChildren
		}

		if cctx.Index == 0 {
			// If no columns exist, close tag should follow open tag.
			lc := doc.LastChild()
			if ln, ok := lc.(*GnoColumnNode); !ok || ln.Tag != GnoColumnTagOpen {
				// Mark open tag as wrong and immediately close the context.
				cctx.OpenTag.Error = ErrColumnsInvalidOpenFormat
				node.Error = ErrColumnsUnexpectedOrInvalidTag
			}
		}

		cctx.IsOpen = false
	}

	return node, parser.NoChildren
}

// Continue returns the parser state for continued parsing.
// Not needed in columns context.
func (*columnsParser) Continue(n ast.Node, reader text.Reader, _ parser.Context) parser.State {
	return parser.Close
}

// Close finalizes the parsing of the node.
// Not needed in columns context.
func (*columnsParser) Close(_ ast.Node, reader text.Reader, _ parser.Context) {}

// CanInterruptParagraph determines if the parser can interrupt paragraphs.
func (*columnsParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine checks if the parser can handle indented lines.
func (*columnsParser) CanAcceptIndentedLine() bool {
	return false
}

// columnsRendererHTML implements NodeRenderer.
type columnsRendererHTML struct{}

// RegisterFuncs adds AST objects to the Renderer.
func (r *columnsRendererHTML) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGnoColumn, renderGnoColumns)
}

// renderGnoColumns renders the column node.
func renderGnoColumns(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	cnode, ok := node.(*GnoColumnNode)
	if !ok || !entering {
		return ast.WalkContinue, nil
	}

	// Check for any error
	if err := cnode.Error; err != nil {
		switch {
		case errors.Is(err, ErrColumnsUnexpectedOrInvalidTag):
			fmt.Fprintf(w, "<!-- unexpected/invalid %q omitted -->\n", cnode.String())
		case errors.Is(err, ErrColumnsInvalidOpenFormat):
			fmt.Fprintln(w, "<!-- gno-columns error: open tag should be followed by heading separator or closing tag -->")
		default:
			fmt.Fprintf(w, "<!-- gno-columns error: %s -->\n", err.Error())
		}

		return ast.WalkContinue, nil
	}

	// Render the node
	switch cnode.Tag {
	case GnoColumnTagOpen:
		fmt.Fprintln(w, `<div class="gno-columns">`)

	case GnoColumnTagSep:
		if cnode.Index > 1 {
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintf(w, "<!-- Column %d -->\n", cnode.Index)
		fmt.Fprintln(w, "<div>")

	case GnoColumnTagClose:
		if cnode.Index > 0 {
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintln(w, "</div> <!-- </gno-columns> -->")

	default:
		panic("invalid column tag - should not happen")
	}

	return ast.WalkContinue, nil
}

// columnsASTTransformer implements ASTTransformer.
type columnsASTTransformer struct{}

// Transform modifies the AST to handle unfinished open tags.
func (a *columnsASTTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	// Retrieve the last column context
	cctx, ok := pc.Get(columnContextKey).(*columnsContext)
	if !ok {
		return
	}

	// Check for unclosed contexts.
	if cctx.IsOpen {
		// Ensure the column index is greater than zero
		if cctx.Index == 0 {
			cctx.OpenTag.Error = ErrColumnsInvalidOpenFormat
			return
		}

		// Insert a closing column tag after the last child of the doc
		lc := doc.LastChild()
		nodeCol := NewColumn(cctx, cctx.Index, GnoColumnTagClose)
		doc.InsertAfter(doc, lc, nodeCol)
	}
}

type columns struct{}

// Columns instance for extending markdown with column functionality.
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
		util.Prioritized(&columnsRendererHTML{}, 500),
	))
}
