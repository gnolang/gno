package markdown

import (
	"bytes"
	"html"
	"strings"

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

// MaxMathPageOutput is the maximum number of bytes of MathML a single render
// may produce. Once it is spent, the remaining expressions are rendered as
// escaped text without being converted. MaxMathInputLen bounds one
// expression, this bounds a page made of many.
const MaxMathPageOutput = 2 << 20

// maxMathOutputLen bounds the MathML of one expression of texLen bytes; an
// expression that expands more (a table of thousands of tiny cells) is
// rendered as escaped text.
func maxMathOutputLen(texLen int) int { return 64*texLen + 4096 }

// mathBudget is the MathML output left for one render. Parsers attach the
// render's budget, kept on the parser context, to every math node.
type mathBudget struct{ left int }

var mathBudgetKey = parser.NewContextKey()

func mathBudgetFrom(pc parser.Context) *mathBudget {
	if b, ok := pc.Get(mathBudgetKey).(*mathBudget); ok {
		return b
	}
	b := &mathBudget{left: MaxMathPageOutput}
	pc.Set(mathBudgetKey, b)
	return b
}

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
	budget *mathBudget
}

type mathBlockNode struct {
	ast.BaseBlock
	flavor   int
	tex      string
	openTag  []byte
	closeTag []byte
	closed   bool // the closing delimiter was found
	budget   *mathBudget
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

func (p *texInlineRegionParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
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
	find := func(b []byte) int { return bytes.Index(b, end) }
	if flavor == flavor_inline|delimeter_tex {
		find = findDollarClose
	}
	key := inlineCloseKeys[string(end)]
	start := seg.Start + len(begin)
	stop := findCloseCached(pc, key, line[len(begin):], start, seg.Stop, find)
	if stop < 0 {
		// could be a linebreak due to formatting issues
		posLine, posSeg := block.Position()
		block.AdvanceLine()
		line, seg = block.PeekLine()
		stop = findCloseCached(pc, key, line, seg.Start-seg.Padding, seg.Stop, find)
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
	return &mathInlineNode{tex: tex, flavor: flavor, budget: mathBudgetFrom(pc)}
}

var inlineCloseKeys = map[string]parser.ContextKey{
	string(_dollarInline):  parser.NewContextKey(),
	string(_dollarDisplay): parser.NewContextKey(),
	string(_inlineclose):   parser.NewContextKey(),
	string(_displayclose):  parser.NewContextKey(),
}

// closeScan records one search for a closing delimiter on the line that ends
// at source offset stop: searching from offset from, the first close is at
// offset found, or there is none before the line end if found is -1.
type closeScan struct{ stop, from, found int }

// findCloseCached returns find(b), where b is the part of a line starting at
// source offset base and ending at stop, reusing an earlier search of the same
// line when it answers the question. Whether a delimiter closes depends only
// on the bytes around it, never on where the search started, so a search from
// an earlier offset that found its close at or after base, or found none,
// answers this one too. Without this, every unclosed opener rescans the rest
// of its line (and the next one), which is quadratic on a line of "$a "
// openers. Two lines are remembered: the opener's and the next one.
func findCloseCached(pc parser.Context, key parser.ContextKey, b []byte, base, stop int, find func([]byte) int) int {
	cache, _ := pc.Get(key).([2]closeScan)
	slot := -1
	for i, sc := range cache {
		if sc.stop != stop {
			continue
		}
		if sc.from <= base && (sc.found < 0 || sc.found >= base) {
			if sc.found < 0 {
				return -1
			}
			return sc.found - base
		}
		slot = i
	}
	if slot < 0 { // evict the entry for the earlier line
		slot = 0
		if cache[1].stop < cache[0].stop {
			slot = 1
		}
	}
	idx := find(b)
	found := -1
	if idx >= 0 {
		found = base + idx
	}
	cache[slot] = closeScan{stop: stop, from: base, found: found}
	pc.Set(key, cache)
	return idx
}

// isEscaped reports whether b[i] is preceded by an odd number of backslashes.
func isEscaped(b []byte, i int) bool {
	n := 0
	for i--; i >= 0 && b[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}

// findDollarClose returns the index of the first $ in b that can close an
// inline $...$ expression, or -1. Following pandoc, a closing $ must not be
// preceded by a space or an escaping backslash nor followed by a digit, so
// "$5 and $10" stays plain text.
func findDollarClose(b []byte) int {
	for i := range b {
		if b[i] != '$' {
			continue
		}
		if i == 0 || util.IsSpace(b[i-1]) || isEscaped(b, i) {
			continue
		}
		if i+1 < len(b) && b[i+1] >= '0' && b[i+1] <= '9' {
			continue
		}
		return i
	}
	return -1
}

func (p *texBlockRegionParser) Trigger() []byte {
	return []byte{'\\', '$'}
}

func (p *texBlockRegionParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	if _, ok := parent.(*mathInlineNode); ok {
		return nil, parser.NoChildren
	}

	// Only display delimiters ($$ and \\[) open a math block. Anything else
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
	if !hasClosingLine(reader, pc, closeTag) {
		return nil, parser.NoChildren
	}

	reader.Advance(len(open))
	node := &mathBlockNode{flavor: flavor, openTag: open, closeTag: closeTag, budget: mathBudgetFrom(pc)}
	_, seg := reader.PeekLine()
	node.Lines().Append(seg)
	return node, parser.NoChildren
}

var (
	mathScanDisplayKey = parser.NewContextKey()
	mathScanDollarKey  = parser.NewContextKey()
)

// mathScan caches a closing-delimiter lookahead over source offsets: no line
// starting in [from, to) closes the block or ends it early. If found, the line
// starting at to closes it; if dead, that line ends any block before it (see
// endsMath) or to is the end of the document; otherwise scanning stopped at
// the size limit and can resume from to.
type mathScan struct {
	from, to    int
	found, dead bool
}

// closingLine reports whether line closes a math block delimited by closeTag:
// it must start with the delimiter, or end with it and hold it only once (so
// an inline $$y$$ at the end of a line does not count). It returns the length
// of the content before the delimiter and the number of bytes to consume.
func closingLine(line, closeTag []byte) (content, consumed int, ok bool) {
	trimmed := util.TrimLeftSpace(line)
	if bytes.HasPrefix(trimmed, closeTag) {
		indent := len(line) - len(trimmed)
		return 0, indent + len(closeTag), true
	}
	trimmed = util.TrimRightSpace(line)
	if bytes.HasSuffix(trimmed, closeTag) && bytes.Count(trimmed, closeTag) == 1 {
		content = len(trimmed) - len(closeTag)
		return content, len(trimmed), true
	}
	return 0, 0, false
}

// endsMath reports whether line cannot be part of display math and ends any
// block before it: a blank line (TeX forbids paragraph breaks in math mode),
// or a code fence or ATX heading, which an unclosed opener must not swallow.
func endsMath(line []byte) bool {
	if util.IsBlank(line) {
		return true
	}
	trimmed := util.TrimLeftSpace(line)
	if len(line)-len(trimmed) > 3 {
		return false
	}
	if bytes.HasPrefix(trimmed, []byte("```")) || bytes.HasPrefix(trimmed, []byte("~~~")) {
		return true
	}
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	return level >= 1 && level <= 6 && (level == len(trimmed) || util.IsSpace(trimmed[level]))
}

// hasClosingLine reports whether a closing line for closeTag follows the
// current line within MaxMathInputLen bytes, before any line that ends math
// (see endsMath). It reads the source directly and leaves the reader alone.
// Results are cached on pc so that a page full of unclosed openers is scanned
// once overall instead of once per opener.
func hasClosingLine(reader text.Reader, pc parser.Context, closeTag []byte) bool {
	key := mathScanDollarKey
	if bytes.Equal(closeTag, _displayclose) {
		key = mathScanDisplayKey
	}

	src := reader.Source()
	_, seg := reader.PeekLine()
	start := seg.Stop // the current line includes its newline
	limit := start + MaxMathInputLen

	pos := start
	sc, ok := pc.Get(key).(mathScan)
	if ok && sc.from <= start && start <= sc.to {
		if sc.found {
			return sc.to <= limit
		}
		if sc.dead {
			return false
		}
		pos = sc.to
	} else {
		sc = mathScan{from: start}
	}

	defer func() { pc.Set(key, sc) }()
	for {
		sc.to = pos
		if pos > limit {
			return false
		}
		if pos >= len(src) {
			sc.dead = true
			return false
		}
		end := len(src)
		if i := bytes.IndexByte(src[pos:], '\n'); i >= 0 {
			end = pos + i + 1
		}
		line := src[pos:end]
		if _, _, ok := closingLine(line, closeTag); ok {
			sc.found = true
			return true
		}
		if endsMath(line) {
			sc.dead = true
			return false
		}
		pos = end
	}
}

func (p *texBlockRegionParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	n, ok := node.(*mathBlockNode)
	if !ok {
		return parser.Close
	}
	line, seg := reader.PeekLine()
	if content, consumed, ok := closingLine(line, n.closeTag); ok {
		node.Lines().Append(text.NewSegment(seg.Start, seg.Start+content))
		reader.Advance(consumed) // move reader past closing tag
		n.closed = true
		return parser.Close | parser.NoChildren
	}
	// The lookahead in Open saw a closing line first, but it reads raw lines
	// and cannot see container prefixes: inside a blockquote or list item the
	// container may end first. Leave the line to the other block parsers; the
	// unclosed block renders as text.
	if line == nil || endsMath(line) {
		return parser.Close
	}
	node.Lines().Append(seg)
	return parser.Continue | parser.NoChildren
}

func (p *texBlockRegionParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	if n, ok := node.(*mathBlockNode); ok {
		var tex strings.Builder
		for i := range n.Lines().Len() {
			tex.Write(reader.Value(n.Lines().At(i)))
		}
		n.tex = tex.String()
	}
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
	var budget *mathBudget
	switch t := node.(type) {
	case *mathInlineNode:
		flavor = t.flavor
		tex = t.tex
		budget = t.budget
	case *mathBlockNode:
		if !t.closed {
			// The opener never got its closing line (its container ended
			// first): render the source as plain text.
			w.WriteString("<p>")
			w.WriteString(html.EscapeString(string(t.openTag) + t.tex))
			w.WriteString("</p>\n")
			return ast.WalkSkipChildren, nil
		}
		flavor = t.flavor
		tex = t.tex
		budget = t.budget
	default:
		return ast.WalkContinue, nil
	}
	inline := flavor&flavor_inline > 0
	if budget == nil {
		budget = &mathBudget{left: MaxMathPageOutput}
	}

	if len(tex) <= MaxMathInputLen && budget.left > 0 {
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
		ok := err == nil && len(mml) <= maxMathOutputLen(len(tex)) && len(mml) <= budget.left
		// Charge the budget even for discarded output: it bounds the
		// conversion work of a render, not only what gets written.
		budget.left = max(budget.left-len(mml), 0)
		if ok {
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
