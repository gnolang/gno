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
	// If we're already in an SVG block, don't open a new block for HTML elements
	if pc.Get(svgInfoKey) != nil {
		// fmt.Println("Open: Already in SVG block, refusing to open new block")
		return nil, parser.NoChildren
	}

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

	// fmt.Printf("Open: Found <gno-svg> tag, processing all content in Open\n")
	// Process ALL content in Open() instead of using Continue
	node := NewSvgNode()

	// Skip the opening <gno-svg> tag
	reader.AdvanceLine()

	// Read all lines until we find </gno-svg>
	for {
		line, segment := reader.PeekLine()
		trimmedLine := util.TrimRightSpace(util.TrimLeftSpace(line))

		// fmt.Printf("Open: processing line: %q, segment: %v\n", string(line), segment)

		// Check for closing tag
		if bytes.Contains(trimmedLine, []byte("</gno-svg>")) {
			// fmt.Println("Open: found closing tag, ending")
			reader.AdvanceLine()
			break
		}

		// Append the line as SVG content
		seg := text.NewSegmentPadding(segment.Start, segment.Stop, segment.Padding)
		seg.ForceNewline = true
		node.Lines().Append(seg)
		reader.AdvanceLine()
	}

	return node, parser.Close | parser.NoChildren // Close immediately, no inline parsing
}

func (b *svgBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	// fmt.Println("Continue: Called")
	// Get context data
	data := pc.Get(svgInfoKey)
	if data == nil {
		// fmt.Println("Continue: No context data, returning Close")
		return parser.Close
	}

	// Read all lines until we find </gno-svg>
	for {
		line, segment := reader.PeekLine()
		trimmedLine := util.TrimRightSpace(util.TrimLeftSpace(line))

		fmt.Printf("Continue: processing line: %q, segment: %v\n", string(line), segment)

		// Skip the opening <gno-svg> tag
		if bytes.Contains(trimmedLine, []byte("<gno-svg")) {
			fmt.Println("Continue: skipping <gno-svg> opening tag")
			reader.AdvanceLine()
			continue
		}

		// Check for closing tag
		if bytes.Contains(trimmedLine, []byte("</gno-svg>")) {
			// fmt.Println("Continue: found closing tag, ending")
			reader.AdvanceLine()
			pc.Set(svgInfoKey, nil) // Clear context
			return parser.Close
		}

		// Append the line as SVG content
		seg := text.NewSegmentPadding(segment.Start, segment.Stop, segment.Padding)
		seg.ForceNewline = true
		// fmt.Printf("Appending segment: %v for line: %q\n", seg, string(line))
		node.Lines().Append(seg)
		reader.AdvanceLine()
	}
}

func (b *svgBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (b *svgBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *svgBlockParser) CanAcceptIndentedLine() bool {
	return true // Accept indented lines to prevent other parsers from taking them
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

	// Write opening object tag
	fmt.Fprint(w, `<object type="image/svg+xml">`)

	// Write the SVG content
	l := node.Lines().Len()
	for i := 0; i < l; i++ {
		line := node.Lines().At(i)
		lineContent := line.Value(source)
		// Debug: print what we're actually rendering
		// fmt.Printf("Rendering line %d: %q\n", i, string(lineContent))
		w.Write(lineContent)
	}

	// Write closing object tag
	fmt.Fprint(w, `</object>`)
	w.WriteByte('\n')

	return ast.WalkContinue, nil
}

type svgExtension struct{}

// Extend adds parsing and rendering options for the Form node
func (e *svgExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(NewSVGParser(), 500)), // Absolute highest priority
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
