// Package markdown — nestdepth helper.
//
// nestdepth tracks the global gno-* nesting depth on goldmark's
// parser.Context, providing a shared cap (4 levels) enforced
// uniformly across every participating gno-* extension. The cap
// applies cross-family — a chain like
//
//	<gno-foreign> > <gno-foreign> > <gno-columns> > <gno-columns>
//
// reaches depth 4 (the deepest <gno-columns> is allowed); a 5th
// nested gno-* opener anywhere underneath is refused and falls
// through to raw HTML, which goldmark safe-mode then strips.
//
// The depth counter is a stack: every participating Open calls
// Push, every Close (and the AST-transformer synth-close path)
// calls Pop. Push returns false at cap without incrementing, so
// the caller can refuse cleanly. Pop clamps at 0 to avoid
// underflow in the face of unbalanced transformer Pops.
//
// The block-count counter (gnoForeignBlockKey) is a separate,
// monotonic per-Convert total maintained directly by the foreign
// parser via pc.Get / pc.Set — see ext_foreign.go. It is not
// stack-shaped, so it does not use this helper.
package markdown

import "github.com/yuin/goldmark/parser"

// MaxGnoNestDepth is the hard cap on nested gno-* blocks across
// the family. Reached when a 4th-level opener has Push'd; a 5th
// opener's Push returns false.
const MaxGnoNestDepth = 4

// gnoNestDepthKey stores the current depth integer on
// parser.Context. Set by Push, read by Get, decremented by Pop.
// The same key is pre-seeded into inner-instance contexts via
// Seed so the cap is global across opaque-body boundaries.
var gnoNestDepthKey = parser.NewContextKey()

// gnoForeignBlockKey stores the monotonic per-Convert count of
// <gno-foreign> openers admitted so far. Maintained directly by
// ext_foreign.go (not via this helper) because its lifecycle is
// not stack-shaped — it is never decremented.
var gnoForeignBlockKey = parser.NewContextKey()

// MaxGnoForeignBlocksPerConvert caps the number of <gno-foreign>
// blocks one Convert call will admit. Beyond this count, foreign
// openers fall through to raw HTML.
const MaxGnoForeignBlocksPerConvert = 100

// Get returns the current gno-* nesting depth on pc. Returns 0
// when the key is absent (a freshly-created parser.Context starts
// at depth 0 without any explicit initialization). Returns 0 if
// the key is present with a non-int value (defensive guard against
// hypothetical misuse — only this package sets the key).
func Get(pc parser.Context) int {
	d, ok := pc.Get(gnoNestDepthKey).(int)
	if !ok {
		return 0
	}
	return d
}

// Push increments the depth counter if below the cap. Returns
// true on success (depth incremented), false if already at the
// cap (no change made). Callers should refuse to open their
// block when Push returns false; goldmark will then try the next
// block parser (typically Type-7 HTML block at priority 900),
// which safe-mode strips.
func Push(pc parser.Context) bool {
	d := Get(pc)
	if d >= MaxGnoNestDepth {
		return false
	}
	pc.Set(gnoNestDepthKey, d+1)
	return true
}

// Pop decrements the depth counter, clamped at 0. Callers in
// Close and the AST transformer's synth-close path should call
// Pop exactly once per successful Push.
func Pop(pc parser.Context) {
	d := Get(pc)
	if d <= 0 {
		return
	}
	pc.Set(gnoNestDepthKey, d-1)
}

// Seed pre-seeds the depth on a fresh inner-instance context.
// Opaque-body extensions (currently <gno-foreign>; <gno-card>
// body when CARD.md ships) read the parse-time depth from their
// AST node (stashed during Open) and call Seed when invoking the
// inner goldmark instance, so the cap stays global across the
// inner/outer boundary.
func Seed(pc parser.Context, depth int) {
	pc.Set(gnoNestDepthKey, depth)
}
