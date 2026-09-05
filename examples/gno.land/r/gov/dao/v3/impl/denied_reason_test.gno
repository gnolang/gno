package impl

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/r/gov/dao/v3/memberstore"
)

// DeniedReason is "execution failed: " + err.Error() (govdao.gno), and that
// error comes from the proposal's executor callback — which any third-party
// realm can supply (see govdao_execute_reject_proposal.txtar, where the
// executor lives in gno.land/r/test/request). Rendered raw it let a failing
// proposal write markdown straight beneath the "PROPOSAL HAS BEEN DENIED"
// line, including a forged heading and vote tally.
//
// Operates on proposalStatus.String directly rather than through a real
// proposal, so it does not disturb the proposal ids and counts that the other
// tests in this package assert against shared realm state.
func TestDeniedReasonCannotForgePageStructure(cur realm, t *testing.T) {
	ps := newProposalStatus([]string{memberstore.T1})
	ps.Denied = true
	ps.DeniedReason = "execution failed: boom\n\n### Stats\n\n- **PROPOSAL HAS BEEN ACCEPTED**\n\n---\n<div>swallow the tally"

	out := ps.String(0, cur)

	uassert.False(t, strings.Contains(out, "\n### Stats"),
		"an injected heading must not survive into the rendered stats block")
	uassert.False(t, strings.Contains(out, "\n- **PROPOSAL HAS BEEN ACCEPTED**"),
		"an injected list item must not forge an acceptance line")
	uassert.False(t, strings.Contains(out, "\n---\n"),
		"an injected horizontal rule must not survive")

	// Emphasis and raw HTML are residue that sanitize.Block would have left
	// live; InlineText escapes both. A reader must not see a bold
	// "ACCEPTED" on a denied proposal, and a <div> must not open an HTML
	// block that swallows the vote tally rendered after it.
	uassert.False(t, strings.Contains(out, "**PROPOSAL HAS BEEN ACCEPTED**"),
		"injected bold text must not forge an acceptance line")
	uassert.False(t, strings.Contains(out, "<div>"),
		"injected raw HTML must not open a block that swallows the tally")

	// Folding is the property InlineText adds over the alternatives: the whole
	// reason must land on the REASON: line, so nothing after a newline can be
	// read as new top-level markdown. Asserted by requiring the payload's LAST
	// segment on the same line as its first.
	reason := out[strings.Index(out, "REASON: "):]
	firstLine := reason[:strings.Index(reason, "\n")]
	uassert.True(t, strings.Contains(firstLine, "swallow the tally"),
		"the entire reason must be folded onto the REASON line")
}

// ...and ordinary error text stays readable. InlineText backslash-escapes
// markdown punctuation, so the raw output carries "Boom\!" — but a
// backslash-escaped punctuation mark renders as the bare character, so a
// reader sees "Boom!". Words and spacing are untouched.
func TestDeniedReasonKeepsPlainTextIntact(cur realm, t *testing.T) {
	ps := newProposalStatus([]string{memberstore.T1})
	ps.Denied = true
	ps.DeniedReason = "execution failed: Boom!"

	uassert.True(t, strings.Contains(ps.String(0, cur), `REASON: execution failed: Boom\!`),
		"a plain denial reason must stay on the REASON line, escaped but readable")
}

// The reason reaches an unauthenticated render path and InlineText costs
// ~6,990 gas/byte, so it must be bounded before it is escaped, not after.
// Asserts the clamp at the call site, not clampField in isolation.
//
// The payload is exclamation marks, not letters, and that choice is the whole
// test. InlineText escapes "!" to "\!" but leaves letters alone, so with a
// letter payload the escaped text is the same length as the raw text and both
// orderings produce identical output — the test would pass either way and
// prove nothing. With punctuation, escaping doubles the length, and the two
// orderings become tellable apart.
func TestDeniedReasonIsClampedBeforeSanitizing(cur realm, t *testing.T) {
	ps := newProposalStatus([]string{memberstore.T1})
	ps.Denied = true
	ps.DeniedReason = strings.Repeat("!", 50000)

	out := ps.String(0, cur)

	// The clamp bounds what goes INTO the escaper, so a full maxRenderedReason
	// characters survive and the escaped output is about twice that. Clamping
	// the escaped text instead would leave only half as many, and the escaper
	// would still have processed all 50,000 characters — exactly the cost the
	// clamp exists to avoid.
	uassert.True(t, strings.Count(out, "!") >= maxRenderedReason,
		"the clamp must bound the escaper's input, not its output")
	uassert.True(t, len(out) < 4*maxRenderedReason,
		"the rendered stats block must still be bounded by the clamp")
	uassert.True(t, strings.Contains(out, "… truncated"),
		"a clamped reason must be marked as cut")
}
