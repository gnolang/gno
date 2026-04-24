package markdown

import (
	"bytes"
	"strconv"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// The details extension adds a neutral collapsible block, rendered as a bare
// <details>/<summary> pair without the alert chrome (icon, colored border)
// that the Alert extension applies.
//
// Syntax:
//
//	:::details Summary text
//	arbitrary **markdown**
//	:::
//
// `:::details[open]` makes the block render in its open state. The closing
// fence is a line containing exactly `:::`.

var (
	detailsFenceOpenPrefix = []byte(":::details")
	detailsOpenFlag        = []byte("[open]")
	detailsFenceClose      = []byte(":::")
)

// DetailsBlock is the container node for a `:::details` block.
type DetailsBlock struct {
	ast.BaseBlock
	// Open indicates the block should render with the `open` attribute.
	Open bool
}

// KindDetailsBlock is the AST kind for DetailsBlock.
var KindDetailsBlock = ast.NewNodeKind("DetailsBlock")

// Kind implements ast.Node.
func (n *DetailsBlock) Kind() ast.NodeKind { return KindDetailsBlock }

// Dump implements ast.Node.
func (n *DetailsBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"open": strconv.FormatBool(n.Open),
	}, nil)
}

// DetailsSummary holds the inline summary of a DetailsBlock.
type DetailsSummary struct {
	ast.BaseBlock
}

// KindDetailsSummary is the AST kind for DetailsSummary.
var KindDetailsSummary = ast.NewNodeKind("DetailsSummary")

// Kind implements ast.Node.
func (n *DetailsSummary) Kind() ast.NodeKind { return KindDetailsSummary }

// Dump implements ast.Node.
func (n *DetailsSummary) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

//--- parser

type detailsParser struct{}

// NewDetailsParser returns a goldmark block parser for `:::details` blocks.
func NewDetailsParser() parser.BlockParser { return &detailsParser{} }

// Trigger implements parser.BlockParser.
func (p *detailsParser) Trigger() []byte { return []byte{':'} }

// parseOpenFence inspects the given line (without leading indentation) and
// returns whether it opens a details block, the `open` flag, and the summary
// byte range inside the line (start, end). When the line does not match,
// matched is false.
func parseOpenFence(line []byte) (open bool, summaryStart, summaryEnd int, matched bool) {
	trimmed := bytes.TrimRight(line, "\r\n")
	if !bytes.HasPrefix(trimmed, detailsFenceOpenPrefix) {
		return false, 0, 0, false
	}
	rest := trimmed[len(detailsFenceOpenPrefix):]
	consumed := len(detailsFenceOpenPrefix)

	if bytes.HasPrefix(rest, detailsOpenFlag) {
		open = true
		rest = rest[len(detailsOpenFlag):]
		consumed += len(detailsOpenFlag)
	}

	// After `:::details` (or `:::details[open]`), the rest of the line must
	// be either empty or begin with a space/tab introducing the summary, so
	// that identifiers like `:::detailsmore` are rejected.
	if len(rest) == 0 {
		return open, 0, 0, true
	}
	if rest[0] != ' ' && rest[0] != '\t' {
		return false, 0, 0, false
	}

	// Trim leading whitespace to locate summary start; trim trailing
	// whitespace for a clean segment.
	i := 0
	for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t') {
		i++
	}
	j := len(rest)
	for j > i && (rest[j-1] == ' ' || rest[j-1] == '\t') {
		j--
	}
	if i == j {
		return open, 0, 0, true
	}
	return open, consumed + i, consumed + j, true
}

// isCloseFence returns true when the given line is exactly the `:::` closing
// fence (trailing whitespace and newline allowed).
func isCloseFence(line []byte) bool {
	return bytes.Equal(bytes.TrimRight(line, " \t\r\n"), detailsFenceClose)
}

// Open implements parser.BlockParser.
func (p *detailsParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()

	// Allow up to 3 leading spaces per CommonMark indentation rules.
	pos := 0
	for pos < len(line) && pos < 4 && line[pos] == ' ' {
		pos++
	}
	if pos == 4 { // 4 spaces is an indented code block
		return nil, parser.NoChildren
	}
	sub := line[pos:]

	open, sStart, sEnd, ok := parseOpenFence(sub)
	if !ok {
		return nil, parser.NoChildren
	}

	block := &DetailsBlock{Open: open}

	if sEnd > sStart {
		absStart := segment.Start + pos + sStart
		absEnd := segment.Start + pos + sEnd
		summarySeg := text.NewSegment(absStart, absEnd)

		segments := text.Segments{}
		segments.Append(summarySeg)

		sum := &DetailsSummary{}
		sum.SetLines(&segments)
		block.AppendChild(block, sum)
	}

	// Consume the whole opening line so its content is not re-parsed.
	reader.Advance(len(line))

	return block, parser.HasChildren
}

// Continue implements parser.BlockParser.
func (p *detailsParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if isCloseFence(line) {
		reader.Advance(len(line))
		return parser.Close
	}
	return parser.Continue | parser.HasChildren
}

// Close implements parser.BlockParser.
func (p *detailsParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

// CanInterruptParagraph implements parser.BlockParser.
func (p *detailsParser) CanInterruptParagraph() bool { return true }

// CanAcceptIndentedLine implements parser.BlockParser.
func (p *detailsParser) CanAcceptIndentedLine() bool { return false }

//--- renderer

const (
	detailsOpenTag        = "<details class=\"gno-details\">\n"
	detailsOpenTagExpand  = "<details class=\"gno-details\" open>\n"
	detailsCloseTag       = "</div>\n</details>\n"
	detailsSummaryOpen    = "<summary><svg><use href=\"#ico-arrow\"></use></svg>"
	detailsSummaryDefault = detailsSummaryOpen + "Details</summary>\n<div>\n"
	detailsSummaryClose   = "</summary>\n<div>\n"
)

// DetailsHTMLRenderer renders DetailsBlock and DetailsSummary nodes.
type DetailsHTMLRenderer struct {
	html.Config
}

// NewDetailsHTMLRenderer returns a NodeRenderer for the details extension.
func NewDetailsHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &DetailsHTMLRenderer{Config: html.NewConfig()}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *DetailsHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindDetailsBlock, r.renderDetailsBlock)
	reg.Register(KindDetailsSummary, r.renderDetailsSummary)
}

func (r *DetailsHTMLRenderer) renderDetailsBlock(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*DetailsBlock)
	if !entering {
		w.WriteString(detailsCloseTag)
		return ast.WalkContinue, nil
	}
	if n.Open {
		w.WriteString(detailsOpenTagExpand)
	} else {
		w.WriteString(detailsOpenTag)
	}
	// No summary: emit a default one so the chevron and body wrapper
	// still apply.
	if _, ok := n.FirstChild().(*DetailsSummary); !ok {
		w.WriteString(detailsSummaryDefault)
	}
	return ast.WalkContinue, nil
}

func (r *DetailsHTMLRenderer) renderDetailsSummary(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString(detailsSummaryOpen)
	} else {
		w.WriteString(detailsSummaryClose)
	}
	return ast.WalkContinue, nil
}

//--- extension

type detailsExtension struct{}

// ExtDetails is the global instance of the details extension.
var ExtDetails = &detailsExtension{}

// Extend adds the details parser and renderer to the Goldmark instance.
func (e *detailsExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewDetailsParser(), 799),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewDetailsHTMLRenderer(), 0),
		),
	)
}
