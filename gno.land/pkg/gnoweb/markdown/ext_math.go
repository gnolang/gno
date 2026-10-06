package markdown

import (
	"bytes"
	"fmt"
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

// maxMathOutputLen bounds the MathML of n bytes of math source: one
// expression of n bytes (one that expands more, such as a table of thousands
// of tiny cells, is rendered as escaped text), and a page of n bytes, however
// many expressions it holds (see mathBudgetFrom).
func maxMathOutputLen(n int) int { return 64*n + 4096 }

// mathBudget is the MathML output left for one render. Parsers attach the
// render's budget, kept on the parser context, to every math node.
type mathBudget struct{ left int }

var mathBudgetKey = parser.NewContextKey()

// mathBudgetFrom returns the render's budget, creating it for a page of
// srcLen bytes on first use. The budget scales with the page: every <math>
// element costs a few hundred bytes of fixed markup, so a page made only of
// tiny expressions ($a$$b$...) would otherwise grow by that much per three
// input bytes, well past the ratio one expression is held to. A page gets the
// MathML one expression of the page's size could produce, and never more
// than MaxMathPageOutput; prose between expressions leaves plenty of room
// for real pages.
func mathBudgetFrom(pc parser.Context, srcLen int) *mathBudget {
	if b, ok := pc.Get(mathBudgetKey).(*mathBudget); ok {
		return b
	}
	b := &mathBudget{left: min(maxMathOutputLen(srcLen), MaxMathPageOutput)}
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

// texBlockRegionParser parses display math blocks. starters holds the block
// parsers that end one (see endsMath).
type texBlockRegionParser struct {
	starters *blockStarters
}

// mathFlavor says how an expression is displayed and which delimiters
// enclose it in the source.
type mathFlavor uint8

const (
	flavorInline mathFlavor = 1 << iota
	flavorDisplay
	delimiterAMS // \\( \\) and \\[ \\]
	delimiterTeX // $ and $$
)

var (
	_inlineopen    = []byte(`\\(`)
	_inlineclose   = []byte(`\\)`)
	_displayopen   = []byte(`\\[`)
	_displayclose  = []byte(`\\]`)
	_dollarInline  = []byte("$")
	_dollarDisplay = []byte("$$")
)

// mathExpr is what the renderer needs from an inline or block math node.
type mathExpr struct {
	flavor mathFlavor
	tex    string
	budget *mathBudget
}

type mathInlineNode struct {
	ast.BaseInline
	mathExpr
}

type mathBlockNode struct {
	ast.BaseBlock
	mathExpr
	closeTag []byte
	closed   bool // the closing delimiter was found
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
	posLine, posSeg := block.Position()
	line, seg := block.PeekLine()
	var begin, end []byte
	var flavor mathFlavor
	var key parser.ContextKey
	if len(line) < len(_inlineopen) {
		return nil
	}
	if line[0] == '$' {
		if line[1] == '$' {
			flavor = flavorDisplay | delimiterTeX
			begin = _dollarDisplay
			end = _dollarDisplay
			key = closeDollarDisplayKey
		} else {
			// Pandoc rule: the opening $ must be followed by a non-space,
			// so prices such as "$ 5" are not math.
			if util.IsSpace(line[1]) {
				return nil
			}
			flavor = flavorInline | delimiterTeX
			begin = _dollarInline
			end = _dollarInline
			key = closeDollarInlineKey
		}
	} else {
		switch string(line[:3]) {
		case string(_inlineopen):
			flavor = flavorInline | delimiterAMS
			begin = _inlineopen
			end = _inlineclose
			key = closeInlineKey
		case string(_displayopen):
			flavor = flavorDisplay | delimiterAMS
			begin = _displayopen
			end = _displayclose
			key = closeDisplayKey
		default:
			return nil
		}
	}
	find := func(b []byte) int { return bytes.Index(b, end) }
	if flavor == flavorInline|delimiterTeX {
		find = findDollarClose
	}
	start := seg.Start + len(begin)
	stop := findCloseCached(pc, key, line[len(begin):], start, seg.Stop, find)
	if stop < 0 {
		// could be a linebreak due to formatting issues
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
	value := block.Value(seg)
	if util.IsBlank(value) {
		// An empty expression ($$$$, \\(\\)) holds no math, and converting
		// it would cost a whole <math> element for a few input bytes: leave
		// it as text.
		block.SetPosition(posLine, posSeg)
		return nil
	}
	block.Advance(stop + len(end))
	return &mathInlineNode{mathExpr: mathExpr{tex: string(value), flavor: flavor, budget: mathBudgetFrom(pc, len(block.Source()))}}
}

var (
	closeDollarInlineKey  = parser.NewContextKey()
	closeDollarDisplayKey = parser.NewContextKey()
	closeInlineKey        = parser.NewContextKey()
	closeDisplayKey       = parser.NewContextKey()
)

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
	cache, _ := pc.Get(key).(*[2]closeScan)
	if cache == nil {
		cache = new([2]closeScan)
		pc.Set(key, cache)
	}
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
	var flavor mathFlavor
	switch {
	case bytes.HasPrefix(line, _displayopen):
		open, closeTag, flavor = _displayopen, _displayclose, flavorDisplay|delimiterAMS
	case bytes.HasPrefix(line, _dollarDisplay):
		open, closeTag, flavor = _dollarDisplay, _dollarDisplay, flavorDisplay|delimiterTeX
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
	if !p.hasClosingLine(reader, pc, closeTag) {
		return nil, parser.NoChildren
	}

	reader.Advance(len(open))
	node := &mathBlockNode{mathExpr: mathExpr{flavor: flavor, budget: mathBudgetFrom(pc, len(reader.Source()))}, closeTag: closeTag}
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
// the delimiter must be the last thing on the line, after the math or alone,
// and appear on it only once. Like a closing code fence, it may be followed
// only by spaces: "$$y$$ trailing" or "$$ and more" is not a closing line,
// which would leave the text after it as a paragraph (with a stray $$), and
// an inline $$y$$ at the end of a line does not count either. It returns the
// length of the content before the delimiter and the number of bytes to
// consume.
func closingLine(line, closeTag []byte) (content, consumed int, ok bool) {
	trimmed := util.TrimRightSpace(line)
	if !bytes.HasSuffix(trimmed, closeTag) || bytes.Count(trimmed, closeTag) != 1 {
		return 0, 0, false
	}
	content = len(trimmed) - len(closeTag)
	if util.IsBlank(trimmed[:content]) {
		content = 0 // an indented delimiter alone on its line
	}
	return content, len(trimmed), true
}

// blockStarters indexes the block parsers that can interrupt a paragraph by
// the bytes that trigger them, so that a line is offered only to the parsers
// its first byte can open. It is built once per extension: probing every
// parser on every line of display math cost a Trigger slice per parser.
type blockStarters [256][]parser.BlockParser

// defaultBlockStarters holds the CommonMark parsers only.
var defaultBlockStarters = newBlockStarters()

// newBlockStarters indexes the default CommonMark block parsers and extra,
// block parsers as registered with parser.WithBlockParsers, keeping those
// that can interrupt a paragraph.
func newBlockStarters(extra ...util.PrioritizedValue) *blockStarters {
	var t blockStarters
	for _, v := range append(parser.DefaultBlockParsers(), extra...) {
		bp, ok := v.Value.(parser.BlockParser)
		if !ok || !bp.CanInterruptParagraph() {
			continue
		}
		if _, ok := bp.(*texBlockRegionParser); ok {
			continue // math does not end math
		}
		trig := bp.Trigger()
		if trig == nil { // triggered by any byte
			for c := range t {
				t[c] = append(t[c], bp)
			}
			continue
		}
		for _, c := range trig {
			t[c] = append(t[c], bp)
		}
	}
	return &t
}

// blockProbe is a scratch parser context in which a line can be offered to
// block parsers as if it followed a paragraph, without touching the real
// parse.
type blockProbe struct {
	pc  parser.Context
	doc ast.Node
}

var blockProbeKey = parser.NewContextKey()

func blockProbeFrom(pc parser.Context) *blockProbe {
	if p, ok := pc.Get(blockProbeKey).(*blockProbe); ok {
		return p
	}
	doc := ast.NewDocument()
	para := ast.NewParagraph()
	doc.AppendChild(doc, para)
	probe := &blockProbe{pc: parser.NewContext(), doc: doc}
	probe.pc.SetOpenedBlocks([]parser.Block{{Node: para}})
	pc.Set(blockProbeKey, probe)
	return probe
}

// startsBlock reports whether line would open a block that interrupts a
// paragraph: a thematic break, ATX or setext heading, code fence, list item,
// HTML block, one of the gnoweb blocks the extension was built with (columns,
// forms, alerts) or, with quotes, a blockquote. It asks the parsers
// themselves, so every rule applies (an ordered list must start at 1, an
// empty item does not interrupt, a line indented by four spaces never starts
// one of these blocks).
func (t *blockStarters) startsBlock(pc parser.Context, line []byte, quotes bool) bool {
	w, pos := util.IndentWidth(line, 0)
	if w > 3 || pos >= len(line) {
		return false
	}
	c := line[pos]
	if c == '>' && !quotes {
		return false // the quote (and alert) parsers, see endsMath
	}
	bps := t[c]
	if len(bps) == 0 {
		return false
	}
	probe := blockProbeFrom(pc)
	probe.pc.SetBlockOffset(pos)
	probe.pc.SetBlockIndent(w)
	// Some gnoweb parsers count their nesting depth on the context when
	// they open, and a probe never closes what it opens: start each probe
	// at depth 0 so that the count cannot fill up and make them decline.
	Seed(probe.pc, 0)
	for _, bp := range bps {
		if node, _ := bp.Open(probe.doc, text.NewReader(line), probe.pc); node != nil {
			return true
		}
	}
	return false
}

// endsMath reports whether line cannot be part of display math and ends any
// block before it: a blank line (TeX forbids paragraph breaks in math mode)
// or a line that would interrupt a paragraph (see startsBlock). Display math
// is read like a paragraph so that an opener left unclosed cannot swallow
// the lists, quotes, headings or HTML after it when a later line happens to
// hold $$.
//
// The trade-off: a line of math that reads as a block start, such as "- x",
// "+ 2y", "> 0" or "1. a", ends the block, which is then shown as text, just
// as it would end a paragraph. Write such a line as "{}- x", move the
// operator to the end of the previous line, or indent it by four spaces.
//
// quotes is false for the lookahead in Open, which reads raw source lines:
// inside a blockquote every line starts with ">". Continue sees lines with
// their container prefixes removed and catches a quote there.
func (t *blockStarters) endsMath(pc parser.Context, line []byte, quotes bool) bool {
	return util.IsBlank(line) || t.startsBlock(pc, line, quotes)
}

// hasClosingLine reports whether a closing line for closeTag follows the
// current line within MaxMathInputLen bytes, before any line that ends math
// (see endsMath). It reads the source directly and leaves the reader alone.
// Results are cached on pc so that a page full of unclosed openers is scanned
// once overall instead of once per opener.
func (p *texBlockRegionParser) hasClosingLine(reader text.Reader, pc parser.Context, closeTag []byte) bool {
	key := mathScanDollarKey
	if bytes.Equal(closeTag, _displayclose) {
		key = mathScanDisplayKey
	}

	src := reader.Source()
	_, seg := reader.PeekLine()
	start := seg.Stop // the current line includes its newline
	limit := start + MaxMathInputLen

	// The zero mathScan is a scan from offset 0 that has not started.
	sc, _ := pc.Get(key).(*mathScan)
	if sc == nil {
		sc = new(mathScan)
		pc.Set(key, sc)
	}
	pos := start
	if sc.from <= start && start <= sc.to {
		if sc.found {
			return sc.to <= limit
		}
		if sc.dead {
			return false
		}
		pos = sc.to
	} else {
		*sc = mathScan{from: start}
	}

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
		if p.starters.endsMath(pc, line, false) {
			sc.dead = true
			return false
		}
		if _, _, ok := closingLine(line, closeTag); ok {
			sc.found = true
			return true
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
	// The lookahead in Open saw a closing line first, but it reads raw lines
	// and cannot see container prefixes: inside a blockquote or list item the
	// container may end first, and a quote is only caught here. Leave the
	// line to the other block parsers; the unclosed block renders as text.
	if line == nil || p.starters.endsMath(pc, line, true) {
		return parser.Close
	}
	if content, consumed, ok := closingLine(line, n.closeTag); ok {
		node.Lines().Append(text.NewSegment(seg.Start, seg.Start+content))
		reader.Advance(consumed) // move reader past closing tag
		n.closed = true
		return parser.Close | parser.NoChildren
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
	var expr mathExpr
	switch t := node.(type) {
	case *mathInlineNode:
		expr = t.mathExpr
	case *mathBlockNode:
		// Trim the bytes util.IsBlank counts as space, without copying.
		if !t.closed || strings.Trim(t.tex, " \t\n\r") == "" {
			// The opener never got its closing line (its container ended
			// first), or the block holds no math: render the source as
			// plain text.
			w.WriteString("<p>")
			open := _dollarDisplay
			if t.flavor&delimiterAMS > 0 {
				open = _displayopen
			}
			w.Write(open)
			w.WriteString(html.EscapeString(t.tex))
			if t.closed {
				w.Write(t.closeTag)
			}
			w.WriteString("</p>\n")
			return ast.WalkSkipChildren, nil
		}
		expr = t.mathExpr
	default:
		return ast.WalkContinue, nil
	}
	inline := expr.flavor&flavorInline > 0

	if len(expr.tex) <= MaxMathInputLen && expr.budget.left > 0 {
		// The converter keeps per-expression state, so it must not be shared
		// across concurrent renders.
		converter := mathml.NewMathMLConverter()
		var mml string
		var err error
		if inline {
			mml, err = converter.ConvertInline(expr.tex)
		} else {
			mml, err = converter.ConvertDisplay(expr.tex)
		}
		ok := err == nil && len(mml) <= maxMathOutputLen(len(expr.tex)) && len(mml) <= expr.budget.left
		// Charge the budget even for discarded output: it bounds the
		// conversion work of a render, not only what gets written.
		expr.budget.left = max(expr.budget.left-len(mml), 0)
		if ok {
			w.WriteString(mml)
			return ast.WalkSkipChildren, nil
		}
	}

	// Fallback to the escaped raw LaTeX if conversion fails. An inline
	// node sits inside a paragraph, where a <div> is invalid (the browser
	// closes the paragraph before it), so display math written inline
	// ("text $$x$$ text") falls back to a <span>, which the stylesheet
	// shows as a block like the <div>.
	class := "math-display"
	if inline {
		class = "math-inline"
	}
	tag := "div"
	if node.Kind() == KindMathInline {
		tag = "span"
	}
	fmt.Fprintf(w, `<%s class="%s">%s</%s>`, tag, class, html.EscapeString(expr.tex), tag)
	return ast.WalkSkipChildren, nil
}

type mathMLExtension struct {
	starters *blockStarters
}

// NewExtMath returns a math extension. Besides the CommonMark block starts,
// display math ends at a line that one of blockParsers, as registered with
// parser.WithBlockParsers, would open to interrupt a paragraph: pass the
// block parsers of the other extensions loaded alongside it, so that an
// unclosed $$ cannot swallow their blocks either.
func NewExtMath(blockParsers ...util.PrioritizedValue) goldmark.Extender {
	starters := defaultBlockStarters
	if len(blockParsers) > 0 {
		starters = newBlockStarters(blockParsers...)
	}
	return &mathMLExtension{starters: starters}
}

func (e *mathMLExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewTexInlineRegionParser(), priorityMathInlineParser),
		),
		parser.WithBlockParsers(
			util.Prioritized(&texBlockRegionParser{starters: e.starters}, priorityMathBlockParser),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewMathRenderer(), priorityMathRenderer),
		),
	)
}
