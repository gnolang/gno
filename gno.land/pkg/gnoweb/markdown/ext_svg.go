package markdown

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type svgNode struct {
	ast.BaseBlock
}

var KindGnoSvg = ast.NewNodeKind("GnoSVG")

func (n *svgNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
	return
}

func (n *svgNode) Kind() ast.NodeKind {
	return KindGnoSvg
}

func NewSvgNode() ast.Node {
	return &svgNode{}
}

type svgBlockParser struct {
}

var svgInfoKey = parser.NewContextKey()

var defaultSVGParser = &svgBlockParser{}

// NewFencedCodeBlockParser returns a new BlockParser that
// parses fenced code blocks.
func NewSVGParser() parser.BlockParser {
	return defaultSVGParser
}

type svgData struct {
	char   byte
	indent int
	length int
	node   ast.Node
}

// var svgBlockInfoKey = NewContextKey()

func (b *svgBlockParser) Trigger() []byte {
	return []byte{'<'}
}

func (b *svgBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	line = util.TrimRightSpace(util.TrimLeftSpace(line))
	toks, err := ParseHTMLTokens(bytes.NewReader(line))
	if err != nil || len(toks) != 1 {
		return nil, parser.NoChildren
	}

	tok := toks[0]
	if tok.Data != "gno-svg" {
		return nil, parser.NoChildren
	}

	// if pos < 0 || (line[pos] != '`' && line[pos] != '~') {
	// 	return nil, parser.NoChildren
	// }
	// findent := pos
	// fenceChar := line[pos]
	// i := pos
	// for ; i < len(line) && line[i] == fenceChar; i++ {
	// }
	// oFenceLength := i - pos
	// if oFenceLength < 3 {
	// 	return nil, parser.NoChildren
	// }

	node := NewSvgNode()
	// pc.Set(svgInfoKey, &svgData{fenceChar, findent, oFenceLength, node})
	return node, parser.NoChildren
}

func (b *svgBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	closeLine := util.TrimRightSpace(util.TrimLeftSpace(line))
	if "</gno-svg>" == string(closeLine) {
		reader.AdvanceLine()
		return parser.Close
	}

	// w, pos := util.IndentWidth(line, reader.LineOffset())
	// if w < 4 {
	// 	i := pos
	// 	for ; i < len(line) && line[i] == fdata.char; i++ {
	// 	}
	// 	length := i - pos
	// 	if length >= fdata.length && util.IsBlank(line[i:]) {
	// 		newline := 1
	// 		if line[len(line)-1] != '\n' {
	// 			newline = 0
	// 		}
	// 		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
	// 		return parser.Close
	// 	}
	// }

	pos, padding := util.IndentPosition(line, reader.LineOffset(), segment.Padding)
	if pos < 0 {
		pos = util.FirstNonSpacePosition(line)
		if pos < 0 {
			pos = 0
		}
		padding = 0
	}

	seg := text.NewSegmentPadding(segment.Start+pos, segment.Stop, padding)
	// if code block line starts with a tab, keep a tab as it is.
	if padding != 0 {
		preserveLeadingTabInCodeBlock(&seg, reader, 0)
	}
	seg.ForceNewline = true // EOF as newline
	node.Lines().Append(seg)
	reader.AdvanceLine()
	return parser.Continue | parser.NoChildren
}

func (b *svgBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (b *svgBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *svgBlockParser) CanAcceptIndentedLine() bool {
	return false
}

// svgRenderer renders the Svg node.
// When entering the Svg node, it displays the opening <svg> tag
// and when exiting (after rendering the child inputs),
// it displays the submit button and </svg>.
type svgRenderer struct{}

// NewSvgRenderer creates a new instance of svgRenderer
func NewSvgRenderer() *svgRenderer {
	return &svgRenderer{}
}

// RegisterFuncs registers the render function for the Svg node
func (r *svgRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGnoSvg, r.render)
}

// render renders the Svg node
func (r *svgRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	fmt.Fprintln(w, "</mysvg>")
	l := node.Lines().Len()
	for i := 0; i < l; i++ {
		line := node.Lines().At(i)
		w.Write(line.Value(source))
	}

	fmt.Fprintln(w, "</mysvg>")
	return ast.WalkContinue, nil
}

type svgExtension struct{}

// Extend adds parsing and rendering options for the Form node
func (e *svgExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(NewSVGParser(), 500)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(NewSvgRenderer(), 500)),
	)

}

var ExtSvg = &svgExtension{}

func preserveLeadingTabInCodeBlock(segment *text.Segment, reader text.Reader, indent int) {
	offsetWithPadding := reader.LineOffset() + indent
	sl, ss := reader.Position()
	reader.SetPosition(sl, text.NewSegment(ss.Start-1, ss.Stop))
	if offsetWithPadding == reader.LineOffset() {
		segment.Padding = 0
		segment.Start--
	}
	reader.SetPosition(sl, ss)
}
