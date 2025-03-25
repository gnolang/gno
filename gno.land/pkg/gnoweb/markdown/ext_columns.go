package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

// Error messages for invalid column tags.
var ErrColumnsUnexpectedOrInvalidTag = errors.New("unexpected or invalid tag")

// Define custom node kind and maximum heading level.
var (
	KindColumn = ast.NewNodeKind("Column")
	MaxHeading = 6
)

// ColumnTag represents the type of tag in a column block.
type ColumnTag int

const (
	ColumnTagUndefined ColumnTag = iota

	ColumnTagOpen
	ColumnTagClose

	ColumnTagSepOpen
	ColumnTagSepClose
)

var columnTagNames = map[ColumnTag]string{
	ColumnTagUndefined: "ColumnTagUndefined",

	ColumnTagOpen:  "ColumnTagOpen",
	ColumnTagClose: "ColumnTagClose",

	ColumnTagSepOpen:  "ColumnTagSepOpen",
	ColumnTagSepClose: "ColumnTagSepClose",
}

// // ColumnNode represents a semantic tree for a "column".
type ColumnNode struct {
	ast.BaseBlock
	Index int // Index of the column associated with the node.
	Depth int
	Tag   ColumnTag // Current Column Tag for this node.
	Error error     // If not nil, indicates that the node is invalid.

	ctx *columnsContext
}

// Dump implements Node.Dump for debug representation.
func (n *ColumnNode) Dump(source []byte, level int) {
	kv := map[string]string{
		"tag": columnTagNames[n.Tag],
	}
	if n.Tag == ColumnTagSepOpen {
		kv["index"] = strconv.Itoa(n.Index)
	}
	if err := n.Error; err != nil {
		kv["error"] = err.Error()
	}

	ast.DumpHelper(n, source, level, kv, nil)
}

// Kind implements Node.Kind.
func (*ColumnNode) Kind() ast.NodeKind {
	return KindColumn
}

func (n *ColumnNode) String() string {
	return columnTagNames[n.Tag]
}

// NewColumn initializes a ColumnNode object.
func NewColumn(ctx *columnsContext, tag ColumnTag) *ColumnNode {
	return &ColumnNode{ctx: ctx, Tag: tag}
}

var columnContextKey = parser.NewContextKey()

// columnsContext is used to keep track of columns' state across parsing.
type columnsContext struct {
	PrevContext *columnsContext
	IsOpen      bool        // Indicates if a block has been correctly opened.
	Index       int         // Index of the current column; 0 indicates no column.
	OpenTag     *ColumnNode // First opening tag for this context.
	Depth       int
}

// parseLineTag identifies the tag type based on the line content.
// It returns a ColumnTag and a slice of comments if applicable.
func parseLineTag(line []byte) ColumnTag {
	line = util.TrimRightSpace(util.TrimLeftSpace(line))

	// Parse the line into HTML tokens
	toks, err := ParseHTMLTokens(bytes.NewReader(line))
	if err != nil || len(toks) != 1 {
		return ColumnTagUndefined // Return early if error or no tokens
	}

	var tag ColumnTag

	// Determine tag type based on the first token
	switch tok := toks[0]; tok.Data {
	case "gno-columns":
		switch tok.Type {
		case html.StartTagToken:
			tag = ColumnTagOpen
		case html.EndTagToken:
			tag = ColumnTagClose
		}
	case "col", "column":
		switch tok.Type {
		case html.StartTagToken:
			tag = ColumnTagSepOpen
		case html.EndTagToken:
			tag = ColumnTagSepClose
		}
	}

	return tag
}

// columnsParser implements BlockParser.
var _ parser.BlockParser = (*columnsParser)(nil)

type columnsParser struct{}

// Trigger returns the trigger characters for the parser.
func (*columnsParser) Trigger() []byte {
	return []byte{'<'}
}

// Open creates a column node based on the line tag.
func (p *columnsParser) Open(doc ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	// Columns tag cannot be a child of another node.
	if doc.Parent() != nil {
		return nil, parser.NoChildren
	}

	line, _ := reader.PeekLine()
	tag := parseLineTag(line)
	if tag == ColumnTagUndefined {
		return nil, parser.NoChildren
	}

	// Get column context.
	cctx, ok := pc.Get(columnContextKey).(*columnsContext)
	if !ok || !cctx.IsOpen {
		cctx = &columnsContext{PrevContext: cctx}
		pc.Set(columnContextKey, cctx)
	}

	node := NewColumn(cctx, tag)

	switch tag {
	case ColumnTagOpen:
		if cctx.IsOpen {
			node.Error = ErrColumnsUnexpectedOrInvalidTag
			return node, parser.NoChildren
		}

		cctx.OpenTag = node
		cctx.IsOpen = true

	case ColumnTagClose:
		if !cctx.IsOpen || cctx.Depth > 0 {
			node.Error = ErrColumnsUnexpectedOrInvalidTag
			return node, parser.NoChildren
		}

		cctx.IsOpen = false

	case ColumnTagSepOpen:
		if !cctx.IsOpen {
			node.Error = ErrColumnsUnexpectedOrInvalidTag
			return node, parser.NoChildren
		}

		cctx.Index++
		node.Index = cctx.Index

		node.Depth = cctx.Depth
		cctx.Depth++

	case ColumnTagSepClose:
		if !cctx.IsOpen || cctx.Depth <= 0 {
			node.Error = ErrColumnsUnexpectedOrInvalidTag
			return node, parser.NoChildren
		}

		cctx.Depth--
		node.Depth = cctx.Depth
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
	return true
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
		lc := doc.LastChild()
		for depth := cctx.Depth; depth > 0; depth-- {
			// Insert a closing column tag after the last child of the doc
			nodeCol := NewColumn(cctx, ColumnTagSepClose)
			nodeCol.Depth = depth
			doc.InsertAfter(doc, lc, nodeCol)
			lc = nodeCol
		}

		nodeCol := NewColumn(cctx, ColumnTagClose)
		doc.InsertAfter(doc, lc, nodeCol)
	}

	// Ensure no intermediary content.
	for ; cctx != nil; cctx = cctx.PrevContext {
		var openNode ast.Node = cctx.OpenTag
		var withinColumn bool

		for node := openNode.NextSibling(); node != nil; node = node.NextSibling() {
			col, ok := node.(*ColumnNode)
			if !ok {
				if !withinColumn {
					doc.RemoveChild(node.Parent(), node)
					doc.InsertBefore(doc, openNode, node)
				}

				continue
			}

			if col.Tag == ColumnTagSepOpen || col.Tag == ColumnTagSepClose {
				withinColumn = col.Tag == ColumnTagSepOpen
				continue
			}

			break
		}
	}
}

// columnsRendererHTML implements NodeRenderer.
type columnsRendererHTML struct{}

// RegisterFuncs adds AST objects to the Renderer.
func (r *columnsRendererHTML) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindColumn, columnsRenderHTML)
}

// columnsRenderHTML renders the column node.
func columnsRenderHTML(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	cnode, ok := node.(*ColumnNode)
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
	case ColumnTagOpen:
		fmt.Fprintln(w, `<div class="gno-columns">`)

	case ColumnTagSepOpen:
		indent := strings.Repeat("\t", cnode.Depth)
		fmt.Fprintf(w, indent+"<!-- Column %d -->\n", cnode.Index)

		fmt.Fprintln(w, `<div class="gno-col">`)

	case ColumnTagSepClose:
		fmt.Fprintln(w, "</div>")

	case ColumnTagClose:
		fmt.Fprintln(w, "</div> <!-- </gno-columns> -->")

	default:
		panic("invalid column tag - should not happen")
	}

	return ast.WalkContinue, nil
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
