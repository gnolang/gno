package markdown

import (
	"bytes"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/mathml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

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
	// fmt.Println(string(line))
	start := seg.Start + len(begin)
	stop := bytes.Index(line[len(begin):], end)
	if stop < 0 {
		// could be a linebreak due to formatting issues
		// count := 0
		posLine, posSeg := block.Position()
		block.AdvanceLine()
		line, seg = block.PeekLine()
		stop = bytes.Index(line, end)
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

var mathBlockInfoKey = parser.NewContextKey()

type mathBlockData struct {
	start  int
	end    int
	flavor int
}

func (p *texBlockRegionParser) Trigger() []byte {
	return []byte{'\\', '$'}
}

func (p *texBlockRegionParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	if _, ok := parent.(*mathInlineNode); ok {
		return nil, parser.NoChildren
	}

	line, _ := reader.PeekLine()
	displaystyle := false
	var flavor int
	if bytes.HasPrefix(line, _displayopen) {
		displaystyle = true
		flavor = flavor_display | delimeter_ams
	} else if bytes.HasPrefix(line, _dollarDisplay) {
		displaystyle = true
		flavor = flavor_display | delimeter_tex
	} else if bytes.HasPrefix(line, _inlineopen) {
		flavor = flavor_inline | delimeter_ams
	} else if bytes.HasPrefix(line, _dollarInline) {
		flavor = flavor_inline | delimeter_tex
	}

	if displaystyle {
		if flavor&delimeter_ams > 0 {
			if bytes.Contains(line, _displayclose) {
				return nil, parser.NoChildren
			}
			reader.Advance(len(_displayclose))
		}
		if flavor&delimeter_tex > 0 {
			if bytes.Contains(line[2:], _dollarDisplay) {
				return nil, parser.NoChildren
			}
			reader.Advance(len(_dollarDisplay))
		}
		pc.Set(mathBlockInfoKey, mathBlockData{flavor: flavor})
		node := &mathBlockNode{flavor: flavor}
		_, seg := reader.PeekLine()
		node.Lines().Append(seg)
		return node, parser.NoChildren
	}

	// If the closing delimiter is on the same line THEN THIS IS NOT A BLOCK. IT IS INLINE!!!!
	if flavor&delimeter_ams > 0 {
		if bytes.Contains(line, _inlineclose) {
			return nil, parser.NoChildren
		}
		reader.Advance(len(_inlineopen)) // move reader past this line
	} else if flavor&delimeter_tex > 0 {
		if bytes.Contains(line[1:], _dollarInline) {
			return nil, parser.NoChildren
		}
		reader.Advance(len(_dollarInline)) // move reader past this line
	}
	pc.Set(mathBlockInfoKey, mathBlockData{flavor: flavor})
	return &mathBlockNode{flavor: flavor}, parser.NoChildren
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

func (b *texBlockRegionParser) CanInterruptParagraph() bool { return true }

func (b *texBlockRegionParser) CanAcceptIndentedLine() bool { return true }

func (b *texInlineRegionParser) CanInterruptParagraph() bool { return true }

func (b *texInlineRegionParser) CanAcceptIndentedLine() bool { return true }

type MathRenderer struct {
	converter *mathml.MathMLConverter
}

// NewMathRenderer returns a new MathRenderer.
func NewMathRenderer() renderer.NodeRenderer {
	return &MathRenderer{
		converter: mathml.NewMathMLConverter(),
	}
}

// RegisterFuncs registers the renderer with the Goldmark renderer.
func (r *MathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathInline, r.renderMath)
	reg.Register(KindMathBlock, r.renderMath)
}

func (r *MathRenderer) renderMath(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
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
	if entering {
		var mml string
		var err error
		if flavor&flavor_inline > 0 {
			mml, err = r.converter.ConvertInline(tex)
		} else {
			mml, err = r.converter.ConvertDisplay(tex)
		}
		if err != nil {
			// Fallback to raw LaTeX if conversion fails
			if flavor&flavor_inline > 0 {
				w.WriteString(`<span class="math-inline">`)
				w.WriteString(tex)
				w.WriteString(`</span>`)
			} else {
				w.WriteString(`<div class="math-display">`)
				w.WriteString(tex)
				w.WriteString(`</div>`)
			}
		} else {
			w.WriteString(mml)
		}
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
