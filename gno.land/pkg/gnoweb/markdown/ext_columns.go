package markdown

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

var (
	KindGnoColumn       = ast.NewNodeKind("GnoColumn")
	GnoColumnsShorthand = []byte("|||") // shorthand for column separator
)

// GnoColumnTag represents the type of tag in a column block.
type GnoColumnTag int

const (
	GnoColumnTagUndefined GnoColumnTag = iota

	GnoColumnTagOpen
	GnoColumnTagClose

	GnoColumnTagSep
)

var columnTagNames = map[GnoColumnTag]string{
	GnoColumnTagUndefined: "ColumnTagUndefined",

	GnoColumnTagOpen:  "ColumnTagOpen",
	GnoColumnTagClose: "ColumnTagClose",

	GnoColumnTagSep: "ColumnTagSepClose",
}

// GnoColumnNode represents a semantic tree for a "column".
type GnoColumnNode struct {
	ast.BaseBlock
	Index int          // Index of the column associated with the node.
	Tag   GnoColumnTag // Current Column Tag for this node.

	ctx *columnsContext
}

func UndefinedGnoColumnNode() *GnoColumnNode { return &GnoColumnNode{} }

// Dump implements Node.Dump for debug representation.
func (n *GnoColumnNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"tag": columnTagNames[n.Tag],
	}
	if n.Tag == GnoColumnTagSep {
		kv["index"] = strconv.Itoa(n.Index)
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

// Kind implements Node.Kind.
func (*GnoColumnNode) Kind() ast.NodeKind {
	return KindGnoColumn
}

func (n *GnoColumnNode) String() string {
	return columnTagNames[n.Tag]
}

func (n *GnoColumnNode) IsEmptyColumns() bool {
	ctx := n.ctx
	if ctx.OpenTag == nil {
		return true
	}

	next, ok := ctx.OpenTag.NextSibling().(*GnoColumnNode)
	if ok && next.Tag == GnoColumnTagClose {
		return true
	}

	return false
}

// NewColumn initializes a ColumnNode object.
func NewColumn(ctx *columnsContext, tag GnoColumnTag) *GnoColumnNode {
	return &GnoColumnNode{ctx: ctx, Tag: tag}
}

var columnContextKey = parser.NewContextKey()

// columnsContext is used to keep track of columns' state across parsing.
type columnsContext struct {
	IsOpen  bool           // Indicates if a block has been correctly opened.
	Index   int            // Index of the current column; 0 indicates no column.
	OpenTag *GnoColumnNode // First opening tag for this context.
}

// parseLineTag identifies the tag type based on the line content.
// It returns a ColumnTag and a slice of comments if applicable.
func parseLineTag(line []byte) GnoColumnTag {
	// Check for shorthand ||| separator
	if bytes.Equal(line, GnoColumnsShorthand) {
		return GnoColumnTagSep
	}

	// Parse the line into HTML tokens
	toks, err := ParseHTMLTokens(bytes.NewReader(line))
	if err != nil || len(toks) != 1 {
		return GnoColumnTagUndefined // Return early if error or no tokens
	}

	// Determine tag type based on the first token
	switch tok := toks[0]; tok.Data {
	case "gno-columns":
		switch tok.Type {
		case html.StartTagToken:
			return GnoColumnTagOpen
		case html.EndTagToken:
			return GnoColumnTagClose
		}
	case "gno-columns-sep":
		if tok.Type == html.SelfClosingTagToken {
			return GnoColumnTagSep
		}
	}

	return GnoColumnTagUndefined
}

// columnsParser implements BlockParser.
var _ parser.BlockParser = (*columnsParser)(nil)

type columnsParser struct{}

// Trigger returns the trigger characters for the parser.
func (*columnsParser) Trigger() []byte {
	return []byte{'<', '|'}
}

// Open creates a column node based on the line tag.
func (p *columnsParser) Open(doc ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	// Columns tag cannot be a child of another node.
	if doc.Parent() != nil {
		return nil, parser.NoChildren
	}

	line, _ := reader.PeekLine()
	line = util.TrimRightSpace(util.TrimLeftSpace(line))

	tag := parseLineTag(line)

	if tag == GnoColumnTagUndefined {
		return nil, parser.NoChildren
	}

	// Get column context.
	cctx, ok := pc.Get(columnContextKey).(*columnsContext)
	if !ok || !cctx.IsOpen {
		cctx = &columnsContext{}
		pc.Set(columnContextKey, cctx)
	}

	node := NewColumn(cctx, tag)

	switch tag {
	case GnoColumnTagOpen:
		if cctx.IsOpen {
			node.Tag = GnoColumnTagUndefined
			return node, parser.NoChildren
		}

		cctx.IsOpen = true
		cctx.OpenTag = node

	case GnoColumnTagClose:
		if !cctx.IsOpen {
			node.Tag = GnoColumnTagUndefined
			return node, parser.NoChildren
		}

		cctx.IsOpen = false

	case GnoColumnTagSep:
		if !cctx.IsOpen {
			if bytes.Equal(line, GnoColumnsShorthand) {
				// We return nil to let the parser continue here as we
				// are not in a column context.
				return nil, parser.NoChildren
			}

			node.Tag = GnoColumnTagUndefined
			return node, parser.NoChildren
		}

		cctx.Index++
		node.Index = cctx.Index
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

// columnsASTTransformer implements ASTTransformer.
type columnsASTTransformer struct{}

// Transform modifies the AST to handle unfinished open tags.
func (a *columnsASTTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	// Retrieve the last column context
	cctx, ok := pc.Get(columnContextKey).(*columnsContext)
	if !ok {
		return
	}

	// Check for unclosed tags.
	if cctx.IsOpen {
		nodeCol := NewColumn(cctx, GnoColumnTagClose)
		doc.InsertAfter(doc, doc.LastChild(), nodeCol)
	}
}

// columnsRendererHTML implements NodeRenderer.
type columnsRendererHTML struct{}

// RegisterFuncs adds AST objects to the Renderer.
func (r *columnsRendererHTML) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGnoColumn, renderGnoColumns)
}

// renderGnoColumns renders the column node.
func renderGnoColumns(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	cnode, ok := node.(*GnoColumnNode)
	if !ok {
		return ast.WalkContinue, nil
	}

	// Render the node
	switch cnode.Tag {
	case GnoColumnTagOpen:
		fmt.Fprintln(w, `<div class="gno-columns">`)
		if cnode.IsEmptyColumns() {
			return ast.WalkContinue, nil
		}

		fallthrough // start the first column

	case GnoColumnTagSep:
		if cnode.Index > 0 {
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintf(w, "<!-- Column %d -->\n", cnode.Index)
		fmt.Fprintln(w, `<div class="gno-column">`)

	case GnoColumnTagClose:
		if !cnode.IsEmptyColumns() {
			fmt.Fprintln(w, "</div>")
		}

		fmt.Fprintln(w, "</div> <!-- </gno-columns> -->")

	case GnoColumnTagUndefined:
		fmt.Fprintf(w, "<!-- unexpected/invalid columns tag omitted -->\n")

	default:
		panic("invalid column tag - should not happen")
	}

	return ast.WalkContinue, nil
}

type columns struct{}

// ExtColumns instance for extending markdown with column functionality.
var ExtColumns = &columns{}

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
