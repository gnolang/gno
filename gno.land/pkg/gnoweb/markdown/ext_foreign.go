// Package markdown — gno-foreign extension.
//
// `<gno-foreign>` is a render-time sandbox primitive for foreign-built
// markdown (markdown returned by an interface method, fetched from a
// foreign realm, etc.). The body is collected verbatim by the outer
// parser (opaque to all other block parsers thanks to the
// load-bearing `parser.Continue` (no HasChildren) invariant) and
// rendered inside its own goldmark instance with structural
// extensions selectively loaded.
package markdown

import (
	"bytes"
	"fmt"
	htmlpkg "html"

	chainmd "github.com/gnolang/gno/gnovm/stdlibs/chain/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

// KindGnoForeign is the node kind for ForeignNode.
var KindGnoForeign = ast.NewNodeKind("GnoForeign")

// MaxGnoForeignBlocksPerConvert caps the number of <gno-foreign> blocks
// one Convert call will admit. Beyond this count, foreign openers fall
// through to raw HTML (safe-mode strips them). It is a foreign-specific
// MONOTONIC per-Convert total — never decremented — distinct from the
// cross-family nesting depth (nestdepth.go).
//
// Sourced from the chain/markdown native (the single source of truth)
// so this enforcement and the realm-facing value realms read via
// markdown.MaxForeignBlocksPerConvert() cannot drift apart.
var MaxGnoForeignBlocksPerConvert = chainmd.MaxForeignBlocksPerConvert()

// gnoForeignBlockKey stores the monotonic per-Convert count of
// <gno-foreign> openers admitted so far, maintained directly via
// pc.Get / pc.Set in Open. Its lifecycle is not stack-shaped (never
// decremented), so it does not use the nestdepth Push/Pop helper.
var gnoForeignBlockKey = parser.NewContextKey()

// ForeignNode is the AST node for a `<gno-foreign>` block. The body
// bytes are collected verbatim by the outer parser and handed to an
// inner goldmark instance at render time.
type ForeignNode struct {
	ast.BaseBlock
	// Body is the accumulated verbatim bytes between the opener and
	// the matching close tag (inclusive of any line terminators).
	Body []byte
	// Label is the optional label string (v1: always "external
	// content"; attribute parsing deferred to a future PR).
	Label string
	// DepthAtParse is the gnoNestDepth value AFTER this block's Push
	// fired during Open. The renderer reads it back to seed the
	// inner instance's parser.Context — goldmark's renderer signature
	// does not carry parser.Context, so the value travels via this
	// AST-node field.
	DepthAtParse int
	// GnoCtx is the render context (GnoURL, chain id, …) captured at
	// parse time. The renderer rebuilds the inner instance's
	// parser.Context from it so links inside the sandbox get the same
	// URL-aware, dangerous-URL-guarded treatment as top-level content
	// (an empty context makes the link transformer no-op, leaving
	// autolinks like `<javascript:…>` unsanitized). Travels via the
	// node for the same reason as DepthAtParse.
	GnoCtx GnoContext
	// Closed is true if the outer parser saw a matching close tag
	// before EOF. False if the AST transformer had to synth-close
	// the node. The renderer treats both the same way.
	Closed bool
	// framingDepth is the local body-framing counter — NOT the
	// shared gnoNestDepth. It is incremented for every inner
	// `<gno-foreign>` literal encountered in the body bytes and
	// decremented for every `</gno-foreign>` literal; only when it
	// is zero does a `</gno-foreign>` close the outer block. This
	// keeps nested-looking close tags inside the body from
	// terminating the outer block prematurely.
	framingDepth int
}

// Dump implements ast.Node.
func (n *ForeignNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"depth_at_parse": fmt.Sprintf("%d", n.DepthAtParse),
		"closed":         fmt.Sprintf("%t", n.Closed),
		"body_len":       fmt.Sprintf("%d", len(n.Body)),
	}, nil)
}

// Kind implements ast.Node.
func (*ForeignNode) Kind() ast.NodeKind { return KindGnoForeign }

// ----- tag recognition -----

type foreignTagKind int

const (
	foreignTagNone foreignTagKind = iota
	foreignTagOpen
	foreignTagClose
)

// parseForeignLineTag returns foreignTagOpen / foreignTagClose if the
// line (after trim) is exactly `<gno-foreign>`, `</gno-foreign>`, or
// `<gno-foreign label="…">` (opener only). The label value is
// returned alongside the open kind; close tags and bare openers
// return an empty label.
//
// An opener with any attribute other than `label`, with multiple
// attributes, or any close tag with attributes returns foreignTagNone
// — so unsupported forms fail closed (fall through to raw HTML →
// safe-mode strip) instead of being misinterpreted.
func parseForeignLineTag(line []byte) (foreignTagKind, string) {
	toks, err := ParseHTMLTokens(bytes.NewReader(line))
	if err != nil || len(toks) != 1 {
		return foreignTagNone, ""
	}
	tok := toks[0]
	// CASE-INSENSITIVE by construction: golang.org/x/net/html lowercases
	// tag names, so `<GNO-FOREIGN>` / `<Gno-Foreign>` also reach here as
	// "gno-foreign". This matches the sibling gno-* extensions
	// (ext_columns.go does the same). The realm-side escaper
	// (foreign.isForeignSentinelLine) MUST mirror this case-folding; a
	// case-variant sentinel left unescaped in foreign body bytes would
	// otherwise be recognized here and break the sandbox boundary. Do
	// NOT change this to a case-sensitive compare in isolation — the
	// escaper and parser can't be kept in lockstep across the Gno/Go
	// boundary, so case-insensitive is the safer shared invariant.
	if tok.Data != "gno-foreign" {
		return foreignTagNone, ""
	}

	var label string
	if len(tok.Attr) != 0 {
		// Accept exactly one attribute and only when it is `label`.
		if len(tok.Attr) != 1 || tok.Attr[0].Key != "label" {
			return foreignTagNone, ""
		}
		label = tok.Attr[0].Val
	}

	switch tok.Type {
	case html.StartTagToken:
		return foreignTagOpen, label
	case html.EndTagToken:
		// Close tags with attributes are malformed — reject.
		if label != "" {
			return foreignTagNone, ""
		}
		return foreignTagClose, ""
	}
	return foreignTagNone, ""
}

// trimForeignLine strips 0-3 leading spaces (CM §4.5 indent
// tolerance) and trims trailing ASCII whitespace (matches goldmark's
// util.TrimRightSpace behavior of stripping ' ', '\t', '\n', '\v',
// '\f', '\r').
func trimForeignLine(line []byte) []byte {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	return util.TrimRightSpace(line[i:])
}

// ----- block parser -----

type foreignParser struct{}

var _ parser.BlockParser = (*foreignParser)(nil)

func (*foreignParser) Trigger() []byte { return []byte{'<'} }

func (*foreignParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	kind, label := parseForeignLineTag(trimForeignLine(line))
	if kind != foreignTagOpen {
		return nil, parser.NoChildren
	}

	// Per-Convert block-count cap (foreign-only, monotonic). Checked
	// BEFORE depth so the cap-reached path doesn't perturb the depth
	// counter for sibling extensions.
	//
	// The block-count is MONOTONIC: incremented on every successful
	// Open, NEVER decremented in Close or the AST transformer. It
	// bounds work-per-Convert, not currently-open foreigns.
	blockCount, _ := pc.Get(gnoForeignBlockKey).(int)
	if blockCount >= MaxGnoForeignBlocksPerConvert {
		return nil, parser.NoChildren
	}

	// Cross-family nesting cap.
	depthBefore := Get(pc)
	if !Push(pc) {
		return nil, parser.NoChildren
	}

	pc.Set(gnoForeignBlockKey, blockCount+1)
	reader.AdvanceToEOL()

	// Label stays empty when the opener carries no label attribute; the
	// renderer then omits the label strip entirely (no default text).
	node := &ForeignNode{
		Label:        label,
		DepthAtParse: depthBefore + 1,
		GnoCtx:       getGnoContext(pc),
	}
	// parser.NoChildren — load-bearing opacity invariant: the body must
	// stay opaque to every other block parser.
	return node, parser.NoChildren
}

func (*foreignParser) Continue(n ast.Node, reader text.Reader, pc parser.Context) parser.State {
	fn := n.(*ForeignNode)
	line, _ := reader.PeekLine()
	if len(line) == 0 {
		// EOF: close the block. Mark Closed so the AST transformer's
		// defensive synth-close path is a no-op and doesn't double-Pop
		// the depth counter (goldmark calls parser.Close on every
		// opened block at EOF — see foreignASTTransformer.Transform).
		fn.Closed = true
		return parser.Close
	}

	kind, _ := parseForeignLineTag(trimForeignLine(line))
	switch kind {
	case foreignTagOpen:
		// Inner opener — opaque to outer parsing, increment body
		// framing depth so the matching inner close doesn't close
		// the outer block. Body bytes collected verbatim.
		fn.framingDepth++
		fn.Body = append(fn.Body, line...)
		reader.AdvanceToEOL()
		return parser.Continue | parser.NoChildren

	case foreignTagClose:
		if fn.framingDepth > 0 {
			// Inner close — pairs with an inner open, still inside
			// our body bytes.
			fn.framingDepth--
			fn.Body = append(fn.Body, line...)
			reader.AdvanceToEOL()
			return parser.Continue | parser.NoChildren
		}
		// Outer close — consume the line and close the block.
		fn.Closed = true
		reader.AdvanceToEOL()
		return parser.Close
	}

	// Non-tag line — body content. Collected verbatim including
	// trailing newline. No other block parser sees this line
	// because we return parser.Continue with NoChildren (opacity
	// invariant).
	fn.Body = append(fn.Body, line...)
	reader.AdvanceToEOL()
	return parser.Continue | parser.NoChildren
}

func (*foreignParser) Close(n ast.Node, _ text.Reader, pc parser.Context) {
	// Pop the depth counter unconditionally on Close — this fires
	// only for successfully-opened nodes (Open returned non-nil),
	// so every Close pairs with a prior Push. The AST transformer's
	// synth-close path also calls Pop for nodes that never reached
	// a Close because the outer parser hit EOF.
	Pop(pc)
}

// CanInterruptParagraph: false. Per CM §4.6, Type-7 HTML blocks
// cannot interrupt paragraphs — and `<gno-foreign>` matches Type-7.
// Realm authors MUST emit a blank line before the opener (this is
// Bulletproof Required #2; the helper at p/nt/md/foreign.gno does
// so automatically by writing `\n\n<gno-foreign>\n`).
func (*foreignParser) CanInterruptParagraph() bool { return false }

// CanAcceptIndentedLine: false. The opener must appear at column 0-3.
// 4+ space indent makes the line an indented code block per CM §4.4
// before any block parser's trigger byte is consulted.
func (*foreignParser) CanAcceptIndentedLine() bool { return false }

// ----- AST transformer -----

type foreignASTTransformer struct{}

// Transform walks the document POST-ORDER as defense-in-depth: in
// practice, every successfully-opened ForeignNode reaches
// foreignParser.Close (goldmark calls Close on every opened block at
// EOF), and both the normal-close and EOF-close paths in Continue
// set fn.Closed = true before returning parser.Close. So this loop
// never finds a !Closed node in well-behaved goldmark runs and the
// Pop here is dead code under normal operation. The Pop remains as a
// belt-and-suspenders guard against a hypothetical goldmark contract
// violation; LIFO post-order ensures correct ordering if it ever
// fires.
//
// Block-count counter is NOT decremented — see comment on
// foreignParser.Open.
func (*foreignASTTransformer) Transform(doc *ast.Document, _ text.Reader, pc parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			return ast.WalkContinue, nil
		}
		fn, ok := n.(*ForeignNode)
		if !ok || fn.Closed {
			return ast.WalkContinue, nil
		}
		fn.Closed = true
		Pop(pc)
		return ast.WalkContinue, nil
	})
}

// ----- renderer -----

type foreignRendererHTML struct {
	imgValidator ImageValidatorFunc
}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *foreignRendererHTML) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindGnoForeign, r.renderForeign)
}

func (r *foreignRendererHTML) renderForeign(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n, ok := node.(*ForeignNode)
	if !ok {
		return ast.WalkContinue, nil
	}

	// An empty label renders no label strip and no aria-label — the box
	// is just an unlabeled group. A non-empty label (e.g. from
	// ForeignWithLabel) renders both the visible strip and the
	// accessible name. No "external content" default.
	if label := n.Label; label != "" {
		escLabel := htmlpkg.EscapeString(label)
		fmt.Fprintf(w, "<div class=\"gno-foreign\" role=\"group\" aria-label=\"%s\">\n", escLabel)
		fmt.Fprintf(w, "<div class=\"gno-foreign__label\">%s</div>\n", escLabel)
	} else {
		fmt.Fprintln(w, `<div class="gno-foreign" role="group">`)
	}
	fmt.Fprintln(w, `<div class="gno-foreign__body">`)

	// Inner goldmark instance: built per-render (NOT a package-level
	// singleton) so each foreign block gets isolated parser/renderer
	// state.
	innerGM := buildInnerForeignMarkdown(r.imgValidator)
	// Rebuild the render context (GnoURL, chain id, …) so the inner
	// instance's link transformer runs and applies dangerous-URL
	// guards / rel attributes — without it, autolinks such as
	// `<javascript:…>` render as live hrefs inside the sandbox.
	innerCtx := NewGnoParserContext(n.GnoCtx)
	// Pre-seed the depth counter so the cross-family cap stays
	// global across the inner/outer boundary. gnoForeignBlockKey is
	// intentionally NOT seeded — each Convert maintains its own
	// per-Convert block-count.
	Seed(innerCtx, n.DepthAtParse)
	if err := innerGM.Convert(n.Body, w, parser.WithContext(innerCtx)); err != nil {
		// Inner parse/render failure: surface as a stripped HTML
		// comment rather than aborting the outer render. The body
		// bytes came from foreign-controlled markdown; a parse
		// error here means malformed input, not a bug in the outer
		// pipeline.
		fmt.Fprintf(w, "<!-- gno-foreign inner render error: %s -->",
			htmlpkg.EscapeString(err.Error()))
	}

	fmt.Fprintln(w, "</div>")
	fmt.Fprintln(w, "</div>")
	return ast.WalkSkipChildren, nil
}

// buildInnerForeignMarkdown assembles a goldmark instance for
// rendering opaque `<gno-foreign>` body bytes. Called fresh per
// render (NOT a package-level singleton) so each foreign block
// gets isolated parser/renderer state.
//
// Loaded extensions: the GFM content extensions production gnoweb
// loads (Strikethrough, Table, Footnote, TaskList — see
// render_config.go) so user content renders identically inside the
// sandbox as it would at top level, plus the structural gno-*
// extensions that exist today (foreign, columns, alert) and the link
// extension. Image validator is wired through if non-nil.
//
// The GFM four are content-only (no raw HTML, no scripting); safe mode
// (no WithUnsafe) still strips raw HTML and unrecognized tags like
// <gno-form>. gno-form is intentionally NOT loaded — forms are
// interactive UI and never permitted in foreign-controlled bytes. When
// CARD.md ships, gno-card joins the load list.
//
// Auto-heading-IDs are intentionally NOT enabled here (unlike the outer
// instance): heading anchors inside an opaque sandbox are low-value and
// would multiply cross-block ID collisions.
func buildInnerForeignMarkdown(imgValidator ImageValidatorFunc) goldmark.Markdown {
	m := goldmark.New(
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Table,
			extension.Footnote,
			extension.TaskList,
		),
	)
	ExtForeign.Extend(m, imgValidator) // self — allows nested <gno-foreign>
	ExtColumns.Extend(m)
	ExtAlerts.Extend(m)
	ExtLinks.Extend(m)
	if imgValidator != nil {
		ExtImageValidator.Extend(m, imgValidator)
	}
	return m
}

// ----- extension registration -----

type foreignExtension struct{}

// ExtForeign is the singleton Extender for `<gno-foreign>`. Pass
// the outer goldmark's image validator so the inner instance
// inherits it.
var ExtForeign = &foreignExtension{}

// Extend registers the foreign parser, AST transformer, and
// renderer on the given goldmark instance.
func (e *foreignExtension) Extend(m goldmark.Markdown, imgValidator ImageValidatorFunc) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(&foreignParser{}, 500),
		),
		parser.WithASTTransformers(
			util.Prioritized(&foreignASTTransformer{}, 500),
		),
	)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&foreignRendererHTML{imgValidator: imgValidator}, 500),
	))
}
