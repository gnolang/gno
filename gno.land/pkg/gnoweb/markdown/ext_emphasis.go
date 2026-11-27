// Package markdown — emphasis DoS guard.
//
// goldmark's ProcessDelimiters (parser/delimiter.go) is super-linear in the
// number of '*'/'_' or '~' delimiters in a single block (yuin/goldmark#555):
// a mixed-delimiter payload of ~1 MiB takes tens of seconds. That function is
// package-level and called directly by the core parser, so it cannot be
// replaced via goldmark's public API. Instead we shadow the pluggable emphasis
// InlineParser and cap how many delimiters reach the stack per block, which
// bounds ProcessDelimiters to O(cap^2). Runs beyond the cap render as literal
// text. Under the cap the behavior is identical to goldmark's default parsers.
package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// MaxEmphasisDelimitersPerBlock caps '*', '_' and '~' delimiter runs pushed
// onto the delimiter stack within one block. 1000 runs (500 emphasis spans) in
// a single paragraph is far beyond any real document, while it bounds
// ProcessDelimiters to O(cap^2) — a 1 MiB input drops from ~34 s to
// sub-millisecond.
const MaxEmphasisDelimitersPerBlock = 1000

// gnoEmphasisDelimiterKey holds the per-block count of delimiter runs pushed
// on the parser.Context, '*'/'_' and '~'. Incremented in Parse, reset to 0 in
// CloseBlock, which goldmark calls per block right after ProcessDelimiters.
var gnoEmphasisDelimiterKey = parser.NewContextKey()

// emphasisDelimiterProcessor mirrors goldmark's unexported emphasis delimiter
// processor: '*' and '_' delimiters, same-character pairing, emphasis nodes.
type emphasisDelimiterProcessor struct{}

func (p *emphasisDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '*' || b == '_'
}

func (p *emphasisDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *emphasisDelimiterProcessor) OnMatch(consumes int) ast.Node {
	return ast.NewEmphasis(consumes)
}

var emphasisProcessor = &emphasisDelimiterProcessor{}

// boundedEmphasisParser shadows goldmark's default emphasis parser (registered
// at a lower priority number, so it is tried first). Under the per-block cap it
// behaves identically to the default (scan + push a delimiter); at/over the cap
// it emits the run as literal text so the markers never reach the delimiter
// stack. It implements CloseBlocker to reset the per-block counter; registering
// it via WithInlineParsers auto-registers the CloseBlocker (parser.go:809).
type boundedEmphasisParser struct {
	max int // per-block delimiter-run cap
}

// The per-block reset depends on goldmark discovering CloseBlock via a runtime
// type assertion (parser.go:809); pin both interfaces at compile time so a
// signature drift fails the build instead of silently disabling the reset.
var (
	_ parser.InlineParser = (*boundedEmphasisParser)(nil)
	_ parser.CloseBlocker = (*boundedEmphasisParser)(nil)
)

func (s *boundedEmphasisParser) Trigger() []byte {
	return []byte{'*', '_'}
}

func (s *boundedEmphasisParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, emphasisProcessor)
	if node == nil {
		return nil
	}
	seg := segment.WithStop(segment.Start + node.OriginalLength)

	count, _ := pc.Get(gnoEmphasisDelimiterKey).(int)
	if count >= s.max {
		// Cap reached for this block: consume the run as literal text so it is
		// not pushed as a delimiter. Returning non-nil suppresses the default
		// emphasis parser for this position.
		block.Advance(node.OriginalLength)
		return ast.NewTextSegment(seg)
	}

	node.Segment = seg
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	pc.Set(gnoEmphasisDelimiterKey, count+1)
	return node
}

// strikethroughGuard counts the '~' runs extension.Strikethrough pushes against
// the emphasis cap, and emits them as literal text once it is reached. Without
// that extension the runs still count, so emphasis past the cap stays text.
type strikethroughGuard struct {
	max int
}

var _ parser.InlineParser = (*strikethroughGuard)(nil)

func (s *strikethroughGuard) Trigger() []byte {
	return []byte{'~'}
}

func (s *strikethroughGuard) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	run := util.TrimLeftLength(line, []byte{'~'})
	// The runs extension.Strikethrough pushes: one or two '~' after a non-'~'.
	if run > 2 || before == '~' {
		return nil
	}
	count, _ := pc.Get(gnoEmphasisDelimiterKey).(int)
	if count < s.max {
		pc.Set(gnoEmphasisDelimiterKey, count+1)
		return nil // the strikethrough parser, if loaded, pushes the run
	}
	block.Advance(run)
	return ast.NewTextSegment(segment.WithStop(segment.Start + run))
}

// CloseBlock resets the per-block delimiter counter. goldmark calls this once
// per block, right after ProcessDelimiters, in the closeBlockers loop of
// parseBlocks (parser/parser.go).
func (s *boundedEmphasisParser) CloseBlock(parent ast.Node, block text.Reader, pc parser.Context) {
	pc.Set(gnoEmphasisDelimiterKey, 0)
}

type emphasisGuardExtension struct {
	max int
}

// ExtEmphasis bounds goldmark's emphasis-parsing cost (yuin/goldmark#555). It
// must be registered on every goldmark instance that parses attacker-controlled
// markdown.
var ExtEmphasis = NewExtEmphasis(MaxEmphasisDelimitersPerBlock)

// NewExtEmphasis returns an emphasis guard with the given per-block
// delimiter-run cap, which must be positive — a non-positive cap emits every
// emphasis run as literal text. Production instances use ExtEmphasis; smaller
// caps keep test fixtures (golden files) readable.
func NewExtEmphasis(maxDelimiters int) goldmark.Extender {
	return &emphasisGuardExtension{max: maxDelimiters}
}

// Extend registers the bounded emphasis parser at priority 0 — ahead of every
// other parser on the '*'/'_' trigger bytes (goldmark dispatches ascending;
// its default emphasis parser is 500). The guard owns those bytes: no other
// parser can push an uncounted '*'/'_' delimiter, so the emphasis cap cannot be
// bypassed. A future extension wanting its own '*'/'_' syntax will be starved
// by this guard and must integrate with the cap instead of registering around
// it. The strikethrough guard runs ahead of extension.Strikethrough (500) and
// counts only the '~' runs that parser pushes. A delimiter extension on any
// other byte, or on '~' runs Strikethrough rejects, is uncounted and relies
// on its own parse cost staying linear. Call once per goldmark instance — a
// second call registers a duplicate parser (harmless but wasteful).
func (e *emphasisGuardExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&boundedEmphasisParser{max: e.max}, 0),
			util.Prioritized(&strikethroughGuard{max: e.max}, 0),
		),
	)
}
