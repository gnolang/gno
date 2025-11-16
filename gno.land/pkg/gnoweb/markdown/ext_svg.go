package markdown

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

type svgNode struct {
	ast.BaseBlock
}

func (n *svgNode) IsRaw() bool {
	return true
}

var KindGnoSvg = ast.NewNodeKind("GnoSVG")

func (n *svgNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

func (n *svgNode) Kind() ast.NodeKind {
	return KindGnoSvg
}

func NewSvgNode() ast.Node {
	return &svgNode{}
}

type svgBlockParser struct {
}

var defaultSVGParser = &svgBlockParser{}

// NewSVGParser returns a new BlockParser that parses SVG blocks.
func NewSVGParser() parser.BlockParser {
	return defaultSVGParser
}

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
	if tok.Data != "gno-svg" || tok.Type != html.StartTagToken {
		return nil, parser.NoChildren
	}

	node := NewSvgNode()

	return node, parser.Continue
}

func (b *svgBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	trimmedLine := util.TrimRightSpace(util.TrimLeftSpace(line))

	toks, err := ParseHTMLTokens(bytes.NewReader(trimmedLine))
	if err == nil && len(toks) == 1 {
		tok := toks[0]
		if tok.Data == "gno-svg" && tok.Type == html.EndTagToken {
			reader.AdvanceLine()
			return parser.Close
		}
	}

	// Append the line as SVG content
	seg := text.NewSegmentPadding(segment.Start, segment.Stop, segment.Padding)
	seg.ForceNewline = true
	node.Lines().Append(seg)
	return parser.Continue | parser.NoChildren
}

func (b *svgBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (b *svgBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *svgBlockParser) CanAcceptIndentedLine() bool {
	return true
}

// svgRenderer renders the Svg node.
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

	// Collect SVG content
	var svgContent bytes.Buffer
	for i := range node.Lines().Len() {
		line := node.Lines().At(i)
		lineContent := line.Value(source)
		svgContent.Write(lineContent)
	}

	// Base64 encode the SVG content
	svgData := base64.StdEncoding.EncodeToString(svgContent.Bytes())

	// Write object tag with data URL
	fmt.Fprintf(w, `<object type="image/svg+xml" data="data:image/svg+xml;base64,%s"></object>`, svgData)
	fmt.Fprintln(w)

	return ast.WalkContinue, nil
}

type svgExtension struct{}

// Extend adds parsing and rendering options for the SVG node
func (e *svgExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(NewSVGParser(), 500)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(NewSvgRenderer(), 500)),
	)

}

var ExtSvg = &svgExtension{}
