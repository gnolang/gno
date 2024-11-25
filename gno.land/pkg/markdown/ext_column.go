package markdown

import (
	"fmt"
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

// Define ColumnTag type and constants

type ColumnDirection int

const (
	ColumnDirectionUndefined ColumnDirection = iota
	ColumnDirectionLeft
	ColumnDirectionRight
	ColumnDirectionMiddle
)

func (c ColumnDirection) String() string {
	switch c {
	case ColumnDirectionLeft:
		return "left"
	case ColumnDirectionRight:
		return "right"
	case ColumnDirectionMiddle:
		return "middle"
	default:
		return "undefined"
	}
}

type ColumnTag int

const (
	ColumnTagUndefined ColumnTag = iota
	ColumnTagOpen
	ColumnTagSep
	ColumnTagClose
)

// ColumnNode represents a semantic tree for "column".
type ColumnNode struct {
	ast.BaseBlock
	Index   int
	SepKind byte
	SepDir  ColumnDirection
	Tag     ColumnTag
	ctx     *columnContext
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
	ast.DumpHelper(n, source, level, nil, nil)
}

// Kind implements Node.Kind.
func (n *ColumnNode) Kind() ast.NodeKind {
	return KindColumn
}

// NewColumn initializes a ColumnAST object.
func NewColumn(ctx *columnContext, sepDir ColumnDirection, sepKind byte) *ColumnNode {
	node := &ColumnNode{ctx: ctx, SepDir: sepDir, SepKind: sepKind, Index: 1}
	if ctx != nil {
		node.Index = ctx.index
	}
	return node
}

// columnParser struct and its methods are used for parsing columns.
type columnParser struct{}

var defaultColumnParser = &columnParser{}

func NewColumnBlockParser() parser.BlockParser {
	return defaultColumnParser
}

// Trigger returns the trigger characters for the parser.
func (s *columnParser) Trigger() []byte {
	return []byte{'+', '=', '<', '>'}
}

var dirMap = map[string]ColumnDirection{
	"><":       ColumnDirectionMiddle,
	"<>":       ColumnDirectionMiddle,
	">>":       ColumnDirectionRight,
	"<<":       ColumnDirectionLeft,
	"\x00>":    ColumnDirectionLeft,
	"<\x00":    ColumnDirectionRight,
	"\x00\x00": ColumnDirectionUndefined,
}

// parseSeparator checks if the line starts with any of the given separators and returns the separator kind.
func parseSeparator(line []byte) (byte, ColumnDirection, bool) {
	const minSperatorLen, maxSperatorLen = 3, 255

	line = util.TrimRightSpace(util.TrimLeftSpace(line))
	if len(line) < minSperatorLen || len(line) >= maxSperatorLen {
		return 0, ColumnDirectionUndefined, false
	}

	key := []byte{0, 0}
	if first := line[0]; first == '<' || first == '>' {
		key[0] = first
		line = line[1:]
	}

	lastIndex := len(line) - 1
	if last := line[lastIndex]; last == '<' || last == '>' {
		key[1] = last
		line = line[:lastIndex]
	}

	dir, ok := dirMap[string(key)]
	if !ok {
		return 0, ColumnDirectionUndefined, false
	}

	runes := []byte{'+', '='}
	for _, r := range runes {
		match := true
		for _, c := range line {
			if r != c {
				match = false
				break
			}
		}

		if match {
			return r, dir, true
		}
	}

	return 0, ColumnDirectionUndefined, false
}

var columnContextKey = parser.NewContextKey()

// columnContext struct and its methods are used for handling column context.
type columnContext struct {
	initilized bool
	index      int
}

func (ctx *columnContext) Init() { ctx.initilized = true }

func (ctx *columnContext) Destroy() { ctx.initilized = false }

func (ctx *columnContext) SpanColumn() { ctx.index++ }

func (s *columnParser) getColumnContext(pc parser.Context) *columnContext {
	cctx, ok := pc.Get(columnContextKey).(*columnContext)
	if !ok || !cctx.initilized {
		cctx = &columnContext{}
		pc.Set(columnContextKey, cctx)
	}

	return cctx
}

// Open function opens a new column node based on the separator kind.
func (s *columnParser) Open(_ ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	sepKind, sepDir, ok := parseSeparator(line)
	if !ok {
		return nil, parser.NoChildren
	}

	cctx := s.getColumnContext(pc)
	node := NewColumn(cctx, sepDir, sepKind)
	switch sepKind {
	case '=':
		if !cctx.initilized {
			cctx.Init()
			node.Tag = ColumnTagOpen
		} else {
			cctx.Destroy()
			node.Tag = ColumnTagClose
		}
	case '+':
		if !cctx.initilized {
			return nil, parser.NoChildren
		}

		cctx.SpanColumn()
		node.Tag = ColumnTagSep
	default:
		panic("invalid tag - should not happen")
	}

	reader.Advance(segment.Len())

	node.Index = cctx.index
	return node, parser.NoChildren
}

func (b *columnParser) Continue(_ ast.Node, _ text.Reader, _ parser.Context) parser.State {
	return parser.Close
}

func (b *columnParser) Close(_ ast.Node, _ text.Reader, _ parser.Context) {}

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
			"cols-start",
			fmt.Sprintf("cols-%d", numColumns),
		}

		fmt.Fprintf(w, `<div class="%s">`+"\n", strings.Join(classes, " "))
		fallthrough
	case ColumnTagSep:
		if cnode.Index > 0 {
			fmt.Fprint(w, "</div>\n")
		}

		var classes = []string{
			"col",
		}

		if cnode.SepDir != ColumnDirectionUndefined {
			classes = append(classes, "col-"+cnode.SepDir.String())
		}

		fmt.Fprintf(w, "<!-- Column %d -->\n", cnode.Index+1)
		fmt.Fprintf(w, `<div class="%s">`+"\n", strings.Join(classes, " "))
	case ColumnTagClose:
		fmt.Fprint(w, "</div>\n")
		fmt.Fprint(w, `</div class="cols-end">`+"\n")
	default:
		panic("invalid column tag - should not happend")
	}

	return ast.WalkContinue, nil
}

// column struct is used to extend the markdown with column functionality.
type column struct{}

var Column = &column{}

// Extend function extends the markdown with column functionality.
func (e *column) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(NewColumnBlockParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewRenderer(), 500),
	))
}
