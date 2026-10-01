package markdown

import (
	"bytes"
	"html"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/mathml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// MaxMathInputLen is the maximum length in bytes of a single math expression.
// Longer expressions are not converted and are rendered as escaped text.
const MaxMathInputLen = 8 << 10

const (
	priorityMathInlineParser = 50
	priorityMathBlockParser  = 90
	priorityMathRenderer     = 100
)

type texInlineRegionParser struct{}

func NewTexInlineRegionParser() *texInlineRegionParser {
	return &texInlineRegionParser{}
}

type texBlockRegionParser struct{}

func NewTexBlockRegionParser() *texBlockRegionParser {
	return &texBlockRegionParser{}
}

const (
	flavor_inline = 1 << iota
	flavor_display
	delimeter_ams
	delimeter_tex
)

var (
	_inlineopen    = []byte(`\\(`)
	_inlineclose   = []byte(`\\)`)
	_displayopen   = []byte(`\\[`)
	_displayclose  = []byte(`\\]`)
	_dollarInline  = []byte("$")
	_dollarDisplay = []byte("$$")
)

type mathInlineNode struct {
	ast.BaseInline
	flavor int
	tex    string
}

type mathBlockNode struct {
	ast.BaseBlock
	flavor int
	tex    string
}

var (
	KindMathInline = ast.NewNodeKind("MathInline")
	KindMathBlock  = ast.NewNodeKind("MathBlock")
)

func (n *mathInlineNode) Kind() ast.NodeKind {
	return KindMathInline
}

func (n *mathBlockNode) Kind() ast.NodeKind {
	return KindMathBlock
}

func (n *mathInlineNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

func (n *mathBlockNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

func (p *texInlineRegionParser) Trigger() []byte {
	return []byte{'\\', '$'}
}

func (p *texInlineRegionParser) Parse(parent ast.Node, block text.Reader, _ parser.Context) ast.Node {
	// An escaped delimiter (\$) is literal text.
	if block.PrecendingCharacter() == '\\' {
		return nil
	}
	line, seg := block.PeekLine()
	var begin, end []byte
	var flavor int
	if len(line) < len(_inlineopen) {
		return nil
	}
	if line[0] == '$' {
		if line[1] == '$' {
			flavor = flavor_display | delimeter_tex
			begin = _dollarDisplay
			end = _dollarDisplay
		} else {
			// Pandoc rule: the opening $ must be followed by a non-space,
			// so prices such as "$ 5" are not math.
			if util.IsSpace(line[1]) {
				return nil
			}
			flavor = flavor_inline | delimeter_tex
			begin = _dollarInline
			end = _dollarInline
		}
	} else {
		switch string(line[:3]) {
		case string(_inlineopen):
			flavor = flavor_inline | delimeter_ams
			begin = _inlineopen
			end = _inlineclose
		case string(_displayopen):
			flavor = flavor_display | delimeter_ams
			begin = _displayopen
			end = _displayclose
		default:
			return nil
		}
	}
	findEnd := func(b []byte) int { return bytes.Index(b, end) }
	if flavor == flavor_inline|delimeter_tex {
		findEnd = findDollarClose
	}
	start := seg.Start + len(begin)
	stop := findEnd(line[len(begin):])
	if stop < 0 {
		// could be a linebreak due to formatting issues
		posLine, posSeg := block.Position()
		block.AdvanceLine()
		line, seg = block.PeekLine()
		stop = findEnd(line)
		if stop < 0 {
			block.SetPosition(posLine, posSeg)
			return nil
		}
	} else {
		// there was no linebreak, so we need to account for the slice we took
		// in the original definition of stop.
		stop += len(begin)
	}
	seg = text.NewSegment(start, seg.Start+stop)
	tex := string(block.Value(seg))
	block.Advance(stop + len(end))
	return &mathInlineNode{tex: tex, flavor: flavor}
}

// findDollarClose returns the index of the first $ in b that can close an
// inline $...$ expression, or -1. Following pandoc, a closing $ must not be
// preceded by a space or a backslash nor followed by a digit, so "$5 and $10"
// stays plain text.
func findDollarClose(b []byte) int {
	for i := 0; i < len(b); i++ {
		if b[i] != '$' {
			continue
		}
		if i == 0 || util.IsSpace(b[i-1]) || b[i-1] == '\\' {
			continue
		}
		if i+1 < len(b) && b[i+1] >= '0' && b[i+1] <= '9' {
			continue
		}
		return i
	}
	return -1
}

var mathBlockInfoKey = parser.NewContextKey()

type mathBlockData struct {
	flavor int
}

func (p *texBlockRegionParser) Trigger() []byte {
	return []byte{'\\', '$'}
}

func (p *texBlockRegionParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	if _, ok := parent.(*mathInlineNode); ok {
		return nil, parser.NoChildren
	}

	// Only display delimiters ($$ and \[) open a math block. Anything else
	// (\alpha, \_, $100, ...) is left to the paragraph and inline parsers.
	line, _ := reader.PeekLine()
	var open, closeTag []byte
	var flavor int
	switch {
	case bytes.HasPrefix(line, _displayopen):
		open, closeTag, flavor = _displayopen, _displayclose, flavor_display|delimeter_ams
	case bytes.HasPrefix(line, _dollarDisplay):
		open, closeTag, flavor = _dollarDisplay, _dollarDisplay, flavor_display|delimeter_tex
	default:
		return nil, parser.NoChildren
	}

	// A closing delimiter on the same line is an inline-level expression,
	// handled by the inline parser.
	if bytes.Contains(line[len(open):], closeTag) {
		return nil, parser.NoChildren
	}
	// Don't open a block that never closes: it would swallow the rest of the
	// document.
	if !hasClosingLine(reader, closeTag) {
		return nil, parser.NoChildren
	}

	reader.Advance(len(open))
	pc.Set(mathBlockInfoKey, mathBlockData{flavor: flavor})
	node := &mathBlockNode{flavor: flavor}
	_, seg := reader.PeekLine()
	node.Lines().Append(seg)
	return node, parser.NoChildren
}

// hasClosingLine reports whether closeTag appears on one of the lines after
// the current one, within MaxMathInputLen bytes. The reader position is left
// unchanged.
func hasClosingLine(reader text.Reader, closeTag []byte) bool {
	posLine, posSeg := reader.Position()
	defer reader.SetPosition(posLine, posSeg)
	reader.AdvanceLine()
	for scanned := 0; scanned <= MaxMathInputLen; {
		line, _ := reader.PeekLine()
		if line == nil {
			return false
		}
		if bytes.Contains(line, closeTag) {
			return true
		}
		scanned += len(line)
		reader.AdvanceLine()
	}
	return false
}

func (p *texBlockRegionParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, seg := reader.PeekLine()
	key := pc.Get(mathBlockInfoKey)
	var flavor int
	if d, ok := key.(mathBlockData); ok {
		flavor = d.flavor
	} else {
		return parser.None
	}
	var closeTag []byte
	switch flavor {
	case flavor_inline | delimeter_ams:
		closeTag = _inlineclose
	case flavor_display | delimeter_ams:
		closeTag = _displayclose
	case flavor_inline | delimeter_tex:
		closeTag = _dollarInline
	case flavor_display | delimeter_tex:
		closeTag = _dollarDisplay
	}
	if stop := bytes.Index(line, closeTag); stop > -1 {
		node.Lines().Append(text.NewSegment(seg.Start, seg.Start+stop))
		reader.Advance(stop + len(closeTag)) // move reader past closing tag
		return parser.Close | parser.NoChildren
	}
	node.Lines().Append(seg)
	return parser.Continue | parser.NoChildren
}

func (p *texBlockRegionParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	if d, ok := pc.Get(mathBlockInfoKey).(mathBlockData); ok {
		if n, ok := node.(*mathBlockNode); ok {
			for i := range n.Lines().Len() {
				n.tex += string(reader.Value(n.Lines().At(i)))
			}
			n.flavor = d.flavor
		}
	}
	pc.Set(mathBlockInfoKey, nil)
}

func (p *texBlockRegionParser) CanInterruptParagraph() bool { return true }

func (p *texBlockRegionParser) CanAcceptIndentedLine() bool { return true }

func (p *texInlineRegionParser) CanInterruptParagraph() bool { return true }

func (p *texInlineRegionParser) CanAcceptIndentedLine() bool { return true }

type MathRenderer struct{}

// NewMathRenderer returns a new MathRenderer.
func NewMathRenderer() renderer.NodeRenderer {
	return &MathRenderer{}
}

// RegisterFuncs registers the renderer with the Goldmark renderer.
func (r *MathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathInline, r.renderMath)
	reg.Register(KindMathBlock, r.renderMath)
}

func (r *MathRenderer) renderMath(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	var tex string
	var flavor int
	switch t := node.(type) {
	case *mathInlineNode:
		flavor = t.flavor
		tex = t.tex
	case *mathBlockNode:
		flavor = t.flavor
		tex = t.tex
	default:
		return ast.WalkContinue, nil
	}
	inline := flavor&flavor_inline > 0

	if len(tex) <= MaxMathInputLen {
		// The converter keeps per-expression state, so it must not be shared
		// across concurrent renders.
		converter := mathml.NewMathMLConverter()
		var mml string
		var err error
		if inline {
			mml, err = converter.ConvertInline(tex)
		} else {
			mml, err = converter.ConvertDisplay(tex)
		}
		if err == nil {
			w.WriteString(mml)
			return ast.WalkSkipChildren, nil
		}
	}

	// Fallback to the escaped raw LaTeX if conversion fails.
	if inline {
		w.WriteString(`<span class="math-inline">`)
		w.WriteString(html.EscapeString(tex))
		w.WriteString(`</span>`)
	} else {
		w.WriteString(`<div class="math-display">`)
		w.WriteString(html.EscapeString(tex))
		w.WriteString(`</div>`)
	}
	return ast.WalkSkipChildren, nil
}

type mathMLExtension struct{}

func (e *mathMLExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewTexInlineRegionParser(), priorityMathInlineParser),
		),
		parser.WithBlockParsers(
			util.Prioritized(NewTexBlockRegionParser(), priorityMathBlockParser),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewMathRenderer(), priorityMathRenderer),
		),
	)
}

// ExtMath is the global instance of the math extension
var ExtMath = &mathMLExtension{}
