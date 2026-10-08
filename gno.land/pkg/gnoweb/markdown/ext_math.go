package markdown

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/mathml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
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
	priorityMathTransformer  = 100
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
	openLen  int // the length of the indentation and opening delimiter that start the first line
	closeTag []byte
	closed   bool // the closing delimiter ends the last line
	fence    bool // a ```math fence, held as the only child
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
	find := func(b []byte) (int, int) { return bytes.Index(b, end), len(b) }
	if flavor == flavorInline|delimiterTeX {
		find = findDollarClose
	}
	// The expression is taken from the line slices, which the block reader
	// gives without their container prefix (a quote's > or a list item's
	// indent). block.Value would find the segment by walking the paragraph's
	// lines back from the last one, which is quadratic on a long paragraph
	// of one expression per line.
	first := line[len(begin):]
	var head, tail []byte // the expression on the opener's line and on the next
	stop := findCloseCached(pc, key, first, seg.Start+len(begin), seg.Stop, find)
	if stop >= 0 {
		head = first[:stop]
		stop += len(begin)
	} else {
		// The expression may continue on the next line.
		block.AdvanceLine()
		line, seg = block.PeekLine()
		stop = findCloseCached(pc, key, line, seg.Start-seg.Padding, seg.Stop, find)
		if stop < 0 {
			block.SetPosition(posLine, posSeg)
			return nil
		}
		head, tail = first, line[:stop]
	}
	if util.IsBlank(head) && util.IsBlank(tail) ||
		refersToFootnote(parent, pc, head) || refersToFootnote(parent, pc, tail) {
		// An empty expression ($$$$, \\(\\)) holds no math, and converting
		// it would cost a whole <math> element for a few input bytes: leave
		// it as text. So is one that holds a footnote reference.
		block.SetPosition(posLine, posSeg)
		return nil
	}
	block.Advance(stop + len(end))
	value := string(head) + string(tail)
	node := &mathInlineNode{mathExpr: mathExpr{tex: value, flavor: flavor, budget: mathBudgetFrom(pc, len(block.Source()))}}
	// The expression's source, delimiters included, is its text for
	// whatever reads a node's text rather than rendering it: a heading's
	// table of contents entry, an image's alt text. The renderer skips it.
	node.AppendChild(node, ast.NewString([]byte(string(begin)+value+string(end))))
	return node
}

var footnoteLabelsKey = parser.NewContextKey()

// refersToFootnote reports whether b holds a reference ([^label]) to a
// footnote defined in the document. Such dollars are prices, not math:
// "It costs $5[^1] or 4$ here." would otherwise swallow the reference, and
// goldmark drops a footnote nothing refers to, definition and all. The
// labels are collected once per document, when the first expression holds
// "[^" (block parsing is over by the time inline parsers run).
func refersToFootnote(parent ast.Node, pc parser.Context, b []byte) bool {
	i := bytes.Index(b, []byte("[^"))
	if i < 0 {
		return false
	}
	labels, ok := pc.Get(footnoteLabelsKey).(map[string]bool)
	if !ok {
		labels = map[string]bool{}
		root := parent
		for root.Parent() != nil {
			root = root.Parent()
		}
		_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if fn, ok := n.(*extast.Footnote); ok && entering {
				labels[string(fn.Ref)] = true
			}
			return ast.WalkContinue, nil
		})
		pc.Set(footnoteLabelsKey, labels)
	}
	for ; i >= 0; i = bytes.Index(b, []byte("[^")) {
		b = b[i+2:]
		if j := bytes.IndexByte(b, ']'); j >= 0 && labels[string(b[:j])] {
			return true
		}
	}
	return false
}

var (
	closeDollarInlineKey  = parser.NewContextKey()
	closeDollarDisplayKey = parser.NewContextKey()
	closeInlineKey        = parser.NewContextKey()
	closeDisplayKey       = parser.NewContextKey()
)

// closeScan records one search for a closing delimiter on the line that ends
// at source offset stop: searching from offset from, the first close is at
// offset found, or, if found is -1, there is none before offset end (the
// line end, or a point no expression crosses, see findDollarClose).
type closeScan struct{ stop, from, found, end int }

// closeFinder searches b for a closing delimiter. It returns its index, or
// -1 and the index where the search ended: len(b), or a point no expression
// crosses.
type closeFinder func(b []byte) (idx, end int)

// findCloseCached returns find(b), where b is the part of a line starting at
// source offset base and ending at stop, reusing an earlier search of the same
// line when it answers the question. Whether a delimiter closes depends only
// on the bytes around it, and a point no expression crosses stops every
// search that reaches it, so a search from an earlier offset that found its
// close at or after base, or found none before an end after base, answers
// this one too. Without this, every unclosed opener rescans the rest of its
// line (and the next one), which is quadratic on a line of "$a " openers.
// Two lines are remembered: the opener's and the next one.
func findCloseCached(pc parser.Context, key parser.ContextKey, b []byte, base, stop int, find closeFinder) int {
	if len(b) == 0 {
		// Nothing to search, as past the last line: there is no line to
		// remember either.
		return -1
	}
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
		if sc.from <= base && (sc.found < 0 && base < sc.end || sc.found >= base) {
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
	idx, end := find(b)
	found := -1
	if idx >= 0 {
		found = base + idx
	}
	cache[slot] = closeScan{stop: stop, from: base, found: found, end: base + end}
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
// inline $...$ expression, or -1 and the index where the search ended.
// Following pandoc, a closing $ must not be
// preceded by a space or an escaping backslash nor followed by a digit, so
// "$5 and $10" stays plain text.
//
// Links are kept whole. gnoweb's links to a realm function are written
// [label](/r/x$help&func=F), so a $ inside a link destination, or inside an
// autolink (<scheme:...>), does not close. And an expression that opens in
// a link label ends there: "[Send $10](/r/x$help&func=Send) ... $" has no
// closer, so the link stays a link instead of turning into math.
func findDollarClose(b []byte) (int, int) {
	labels := 0 // [ opened since the start of the expression
	depth := 0  // inside a link destination, its parenthesis nesting
	for i := 0; i < len(b); i++ {
		c := b[i]
		if depth > 0 {
			// isEscaped only on parentheses: on every byte, it would rescan
			// a run of backslashes once per backslash.
			if c == '(' && !isEscaped(b, i) {
				depth++
			} else if c == ')' && !isEscaped(b, i) {
				depth--
			}
			continue
		}
		switch c {
		case '[':
			if !isEscaped(b, i) {
				labels++
			}
		case ']':
			if isEscaped(b, i) {
				continue
			}
			if i+1 < len(b) && b[i+1] == '(' {
				if labels == 0 {
					// The label the expression opened in ends here.
					return -1, i
				}
				depth = 1
				i++
			}
			labels = max(labels-1, 0)
		case '<':
			if end := autolinkEnd(b[i:]); end > 0 {
				i += end
			}
		case '$':
			if i == 0 || util.IsSpace(b[i-1]) || isEscaped(b, i) {
				continue
			}
			if i+1 < len(b) && b[i+1] >= '0' && b[i+1] <= '9' {
				continue
			}
			return i, i
		}
	}
	return -1, len(b)
}

// autolinkEnd returns the index of the > that ends the URI autolink b
// starts with (<scheme:...>, as CommonMark defines it), or -1.
func autolinkEnd(b []byte) int {
	if end := util.FindURLIndex(b[1:]) + 1; end > 0 && end < len(b) && b[end] == '>' {
		return end
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
	// Like a code fence, the opener may be indented by up to three spaces;
	// four make an indented code block.
	line, seg := reader.PeekLine()
	indent := pc.BlockOffset()
	if indent < 0 || pc.BlockIndent() > 3 {
		return nil, parser.NoChildren
	}
	line = line[indent:]
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
	// document. Nor one that holds no math ($$ then $$): like an empty
	// inline expression, it stays text.
	found, empty := p.hasClosingLine(parent, reader, pc, closeTag)
	if !found || empty && util.IsBlank(line[len(open):]) {
		return nil, parser.NoChildren
	}

	node := &mathBlockNode{mathExpr: mathExpr{flavor: flavor, budget: mathBudgetFrom(pc, len(reader.Source()))}, openLen: indent + len(open), closeTag: closeTag}
	node.Lines().Append(seg) // with the delimiter, for a fallback paragraph
	return node, parser.NoChildren
}

var (
	mathScanDisplayKey = parser.NewContextKey()
	mathScanDollarKey  = parser.NewContextKey()
)

// mathScan caches a closing-delimiter lookahead over source offsets, for the
// blocks opened in container: no line starting in [from, to) closes the block
// or ends it early. If found, the line starting at to closes it (and empty
// says it is the first line and holds only the delimiter); if dead, that line
// ends any block before it (see endsMath), leaves the container, or to is the
// end of the document; otherwise scanning stopped at the size limit and can
// resume from to.
type mathScan struct {
	container          ast.Node
	from, to           int
	found, dead, empty bool
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

// probeContext is a parser context whose values cannot be set. Some parsers
// record state for the next line they see (goldmark's list parsers skip the
// next list start, gnoweb's count their nesting depth): a probe never sees
// that next line, and must not leave the state behind for the next probe.
type probeContext struct{ parser.Context }

func (probeContext) Set(parser.ContextKey, any) {}

func (c probeContext) ComputeIfAbsent(key parser.ContextKey, f func() any) any {
	if v := c.Get(key); v != nil {
		return v
	}
	return f()
}

var blockProbeKey = parser.NewContextKey()

func blockProbeFrom(pc parser.Context) *blockProbe {
	if p, ok := pc.Get(blockProbeKey).(*blockProbe); ok {
		return p
	}
	doc := ast.NewDocument()
	para := ast.NewParagraph()
	doc.AppendChild(doc, para)
	probe := &blockProbe{pc: probeContext{parser.NewContext()}, doc: doc}
	probe.pc.SetOpenedBlocks([]parser.Block{{Node: para}})
	pc.Set(blockProbeKey, probe)
	return probe
}

// startsBlock reports whether line would open a block that interrupts a
// paragraph: a thematic break, ATX or setext heading, code fence, list item,
// HTML block, blockquote or one of the blocks of the other extensions the
// math extension was built with (columns, forms, alerts). It asks the parsers
// themselves, so every rule applies (an ordered list must start at 1, an
// empty item does not interrupt, a line indented by four spaces never starts
// one of these blocks).
func (t *blockStarters) startsBlock(pc parser.Context, line []byte) bool {
	w, pos := util.IndentWidth(line, 0)
	if w > 3 || pos >= len(line) {
		return false
	}
	bps := t[line[pos]]
	if len(bps) == 0 {
		return false
	}
	probe := blockProbeFrom(pc)
	probe.pc.SetBlockOffset(pos)
	probe.pc.SetBlockIndent(w)
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
// line is read without its container prefixes (see inContainers).
func (t *blockStarters) endsMath(pc parser.Context, line []byte) bool {
	return util.IsBlank(line) || t.startsBlock(pc, line)
}

// hasClosingLine reports whether a closing line for closeTag follows the
// current line within MaxMathInputLen bytes, before any line that ends math
// (see endsMath) and before the container (blockquote, list item, ...) the
// block would open in ends. It reads the source directly and leaves the
// reader alone. Results are cached on pc so that a page full of unclosed
// openers is scanned once overall instead of once per opener. empty reports
// that the closing line is the next line and holds nothing but the
// delimiter.
func (p *texBlockRegionParser) hasClosingLine(parent ast.Node, reader text.Reader, pc parser.Context, closeTag []byte) (found, empty bool) {
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
	if sc.container == parent && sc.from <= start && start <= sc.to {
		if sc.found {
			return sc.to <= limit, sc.empty && sc.to == start
		}
		if sc.dead {
			return false, false
		}
		pos = sc.to
	} else {
		*sc = mathScan{container: parent, from: start}
	}

	containers := containerBlocks(parent, pc)
	for {
		sc.to = pos
		if pos > limit {
			return false, false
		}
		if pos >= len(src) {
			sc.dead = true
			return false, false
		}
		end := len(src)
		if i := bytes.IndexByte(src[pos:], '\n'); i >= 0 {
			end = pos + i + 1
		}
		line, ok := p.inContainers(pc, containers, src[pos:end])
		if !ok || p.starters.endsMath(pc, line) {
			sc.dead = true
			return false, false
		}
		if content, _, ok := closingLine(line, closeTag); ok {
			sc.found, sc.empty = true, content == 0
			return true, sc.empty && pos == start
		}
		pos = end
	}
}

// containerBlocks returns the open blocks that contain parent, the block a
// math block would open in, outermost first: the containers whose prefix
// each of its lines must carry. Other open blocks, such as the paragraph the
// opener interrupts, are left out.
func containerBlocks(parent ast.Node, pc parser.Context) []parser.Block {
	if parent.Kind() == ast.KindDocument {
		return nil
	}
	var containers []parser.Block
	for _, b := range pc.OpenedBlocks() {
		for n := parent; n != nil; n = n.Parent() {
			if n == b.Node {
				containers = append(containers, b)
				break
			}
		}
	}
	return containers
}

// inContainers offers a source line to the parsers of containers, as the
// parser does to every line before handing what is left to the open block,
// and returns what is left, or false if one of them does not continue on it.
// Display math is not a paragraph, so a line without its container's prefix
// is no lazy continuation line: it ends the container and the block. The
// parsers run in the scratch context of blockProbe, so the real parse is
// left untouched.
func (p *texBlockRegionParser) inContainers(pc parser.Context, containers []parser.Block, line []byte) ([]byte, bool) {
	if len(containers) == 0 {
		return line, true
	}
	probe := blockProbeFrom(pc)
	r := text.NewReader(line)
	for _, c := range containers {
		if c.Parser.Continue(c.Node, r, probe.pc)&parser.Continue == 0 {
			return nil, false
		}
	}
	line, _ = r.PeekLine()
	return line, true
}

func (p *texBlockRegionParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	n, ok := node.(*mathBlockNode)
	if !ok {
		return parser.Close
	}
	line, seg := reader.PeekLine()
	if line == nil || p.starters.endsMath(pc, line) {
		return parser.Close
	}
	if _, consumed, ok := closingLine(line, n.closeTag); ok {
		node.Lines().Append(seg)
		reader.Advance(consumed) // move reader past closing tag
		n.closed = true
		return parser.Close | parser.NoChildren
	}
	node.Lines().Append(seg)
	return parser.Continue | parser.NoChildren
}

// Close reads the TeX between the delimiters. The lookahead in Open checks
// that the closing line comes in the same container and before any line
// that ends math, so the block always closes. Should it not, or should it
// hold no math, it is replaced by a paragraph of the same lines, which the
// inline parsers then read as text, the way they would have without math.
//
// The lines are then trimmed like a paragraph's: the inline parsers read
// them too, for the paragraph the block may become (see mathTransformer).
func (p *texBlockRegionParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	n, ok := node.(*mathBlockNode)
	if !ok {
		return
	}
	lines := n.Lines()
	var tex strings.Builder
	for i := range lines.Len() {
		seg := lines.At(i)
		v := seg.Value(reader.Source())
		if i == 0 {
			v = v[n.openLen:]
		}
		if i == lines.Len()-1 && n.closed {
			content, _, _ := closingLine(v, n.closeTag)
			v = v[:content]
		}
		tex.Write(v)
	}
	n.tex = tex.String()
	for i := range lines.Len() {
		seg := lines.At(i)
		lines.Set(i, seg.TrimLeftSpace(reader.Source()))
	}
	last := lines.At(lines.Len() - 1)
	lines.Set(lines.Len()-1, last.TrimRightSpace(reader.Source()))
	if n.closed && !util.IsBlank([]byte(n.tex)) {
		return
	}
	para := ast.NewParagraph()
	para.SetLines(lines)
	para.SetBlankPreviousLines(n.HasBlankPreviousLines())
	n.Parent().ReplaceChild(n.Parent(), n, para)
}

func (p *texBlockRegionParser) CanInterruptParagraph() bool { return true }

func (p *texBlockRegionParser) CanAcceptIndentedLine() bool { return true }

// mathTransformer finishes the math blocks once the inline parsers are done.
//
// A fenced code block whose info string is exactly "math", GitHub's syntax
// for display math, becomes a display math block, converted with the same
// converter and limits as $$. The code block becomes the math node's only
// child, which the renderer shows, through whatever renders code blocks,
// when the math is not converted: the fence then looks as it would without
// math.
//
// A $$ block whose lines refer to a footnote defined in the document becomes
// the paragraph it would be without math, like an inline expression (see
// refersToFootnote): the block's lines are inline parsed as its children,
// so the reference shows up as a footnote link there.
type mathTransformer struct{}

func (mathTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	src := reader.Source()
	var fences []*ast.FencedCodeBlock
	var withFootnotes []*mathBlockNode
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Type() != ast.TypeBlock {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.FencedCodeBlock:
			if n.Info != nil && string(n.Info.Segment.Value(src)) == "math" {
				fences = append(fences, n)
			}
		case *mathBlockNode:
			if hasFootnoteLink(n) {
				withFootnotes = append(withFootnotes, n)
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	for _, n := range withFootnotes {
		para := ast.NewParagraph()
		para.SetLines(n.Lines())
		para.SetBlankPreviousLines(n.HasBlankPreviousLines())
		for c := n.FirstChild(); c != nil; c = n.FirstChild() {
			para.AppendChild(para, c)
		}
		n.Parent().ReplaceChild(n.Parent(), n, para)
	}
	for _, fence := range fences {
		tex := fence.Lines().Value(src)
		if util.IsBlank(tex) {
			continue // no math, like an empty $$ block
		}
		node := &mathBlockNode{
			mathExpr: mathExpr{tex: string(tex), flavor: flavorDisplay | delimiterTeX, budget: mathBudgetFrom(pc, len(src))},
			closed:   true,
			fence:    true,
		}
		parent := fence.Parent()
		parent.ReplaceChild(parent, fence, node)
		node.AppendChild(node, fence)
	}
}

func hasFootnoteLink(n ast.Node) bool {
	found := false
	_ = ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == extast.KindFootnoteLink {
			found = true
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return found
}

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

	// A ```math fence that is not converted is shown as the code block it
	// holds, as it would be without math.
	if n, ok := node.(*mathBlockNode); ok && n.fence {
		return ast.WalkContinue, nil
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
		parser.WithASTTransformers(
			util.Prioritized(mathTransformer{}, priorityMathTransformer),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewMathRenderer(), priorityMathRenderer),
		),
	)
}
