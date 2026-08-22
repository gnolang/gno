// PKGPATH: gno.land/r/test/disclosure
package disclosure

// Covers the executor-disclosure changes in an isolated realm. The unit tests
// in the impl package share proposal ids across files (govdao_test.gno asserts
// a hard-coded id) and swap the DAO implementation partway through, so these
// live here instead.

import (
	"strings"
	"testing"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

const user address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

func init(cur realm) {
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, user, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))
}

// hostileExecutor implements dao.Executor directly rather than through
// NewSimpleExecutor, which is what lets it choose its own CreationRealm.
// CreationRealm() is dispatched through the public dao.Executor interface, so
// only SimpleExecutor's value is VM-supplied from rlm.PkgPath().
type hostileExecutor struct{}

func (e *hostileExecutor) Execute(cur realm) error { return nil }

func (e *hostileExecutor) String() string { return "" }

func (e *hostileExecutor) CreationRealm() string {
	return "gno.land/r/sys/params\n\n### Stats\n\n- **PROPOSAL HAS BEEN ACCEPTED**\n- YES PERCENT: 100%\n\n---\n"
}

// blankExecutor's CreationRealm is non-empty but strips to nothing:
// sanitize.InlineCode removes bidi and zero-width characters, so it returns "".
type blankExecutor struct{}

func (e *blankExecutor) Execute(cur realm) error { return nil }

func (e *blankExecutor) String() string { return "" }

func (e *blankExecutor) CreationRealm() string { return "\u200b\u200b\u202e" }

// whitespaceExecutor covers the other half of the guard: plain whitespace,
// which InlineCode would otherwise wrap in a padded, empty-looking span.
type whitespaceExecutor struct{}

func (e *whitespaceExecutor) Execute(cur realm) error { return nil }

func (e *whitespaceExecutor) String() string { return "" }

func (e *whitespaceExecutor) CreationRealm() string { return "   \t " }

// hugeExecutor computes a very large CreationRealm while storing nothing.
// Sanitizing costs ~11,310 gas/byte and render is reachable unauthenticated
// under a 3,000,000,000 gas cap, so this must be clamped before escaping.
type hugeExecutor struct{}

func (e *hugeExecutor) Execute(cur realm) error { return nil }

func (e *hugeExecutor) String() string { return "" }

func (e *hugeExecutor) CreationRealm() string { return strings.Repeat("z", 20000) }

// fenceExecutor attacks the code span itself. Section 4 below sends backticks
// through the grant sentence, but that path escapes the value directly. The
// creation realm is clamped first and escaped second, so it needs its own
// case: the fence is chosen after the cut, and must still outscan whatever
// backticks survived it. The run here is two long, so a two-backtick fence
// would be closed by the payload.
type fenceExecutor struct{}

func (e *fenceExecutor) Execute(cur realm) error { return nil }

func (e *fenceExecutor) String() string { return "" }

func (e *fenceExecutor) CreationRealm() string { return "gno.land/r/evil`` **INJECTED**" }

// wsExecutor returns a large run of spaces. TrimSpace used to run before the
// clamp, so it scanned every byte of whatever the executor returned: 250KB of
// spaces cost 1,368,719,824 gas to render nothing at all, and 560KB cost
// 3,048,904,520 — past the query cap, so the page could not be rendered by
// anyone. Clamping first bounds the scan; the same 560KB now costs 16,657,792.
type wsExecutor struct{}

func (e *wsExecutor) Execute(cur realm) error { return nil }

func (e *wsExecutor) String() string { return "" }

func (e *wsExecutor) CreationRealm() string { return strings.Repeat(" ", 20000) }

func main(cur realm) {
	testing.SetOriginCaller(user)

	// 1. An executor with NO description. The creation realm used to share the
	// `ExecutorString() != ""` gate with the description, so it was hidden for
	// every such proposal — 16 call sites across 7 production realms.
	// SetRealm first: NewSimpleExecutor captures rlm.PkgPath(), which is empty
	// outside a code realm.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/template/silent"))
	silent := dao.NewSimpleExecutor(0, cur, func(realm) error { return nil }, "")

	testing.SetRealm(testing.NewUserRealm(user))
	pid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Silent", "A proposal whose executor has no description", silent))
	out := dao.Render(cross(cur), pid.String())

	println("empty-description proposal discloses creation realm:",
		strings.Contains(out, "Executor created in: `gno.land/r/template/silent`"))
	println("and prints no empty metadata block:",
		!strings.Contains(out, "This proposal contains the following metadata"))

	// 2. CreationRealm() is dispatched through the public dao.Executor
	// interface, so a third-party executor picks its own value — and the
	// disclosure now renders for every proposal. InlineCode does not delete the
	// hostile text; it folds it onto one line inside a code span, where it can
	// no longer forge page structure. So assert structure, not absence.
	hpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Hostile", "A proposal whose executor forges page structure", &hostileExecutor{}))
	hout := dao.Render(cross(cur), hpid.String())

	println("only the genuine Stats heading exists:",
		strings.Count(hout, "\n### Stats") == 1)
	println("no forged acceptance line:",
		!strings.Contains(hout, "\n- **PROPOSAL HAS BEEN ACCEPTED**"))
	println("no forged tally line:",
		!strings.Contains(hout, "\n- YES PERCENT: 100%"))
	println("genuine tally and status still render:",
		strings.Contains(hout, "\n- YES PERCENT: 0%") &&
			strings.Contains(hout, "- **Proposal is open for votes**"))
	// One line, inside a code span: the newlines are folded to spaces, so none
	// of the payload can start a block. (No pad space before the value now —
	// TrimSpace removes the payload's trailing newline, so InlineCode does not
	// need to pad the fence.)
	println("hostile value is confined to one code-span line:",
		strings.Contains(hout, "Executor created in: `gno.land/r/sys/params  ### Stats"))

	// 3. The upgrade proposal rewrites AllowedDAOs — the sole authorization for
	// replacing the implementation, mutating the member store and moving
	// treasury funds. It used to carry an empty description, so the realm
	// receiving that authority appeared nowhere a voter would read.
	upid := dao.MustCreateProposal(cross(cur),
		impl.NewUpgradeDaoImplRequest(cross(cur), impl.NewGovDAO(), "gno.land/r/gov/dao/v4/impl", "reason"))
	uout := dao.Render(cross(cur), upid.String())

	println("upgrade proposal states the grant:",
		strings.Contains(uout, "may replace the implementation, mutate the member store, or move treasury funds"))
	println("upgrade proposal names the granted realm:",
		strings.Contains(uout, "`gno.land/r/gov/dao/v4/impl`"))

	// 4. realmPkg is caller-supplied and lands inside a code span. md.EscapeText
	// was wrong there: CommonMark 6.1 does not process backslash escapes inside
	// code spans, so it rendered visible backslashes, and a backtick closed the
	// span early. InlineCode widens the fence instead, so the payload stays
	// inside it as literal text.
	bpid := dao.MustCreateProposal(cross(cur),
		impl.NewUpgradeDaoImplRequest(cross(cur), impl.NewGovDAO(),
			"gno.land/r/x` **PROPOSAL HAS BEEN ACCEPTED** `", "reason"))
	bout := dao.Render(cross(cur), bpid.String())

	// The payload's own backtick sits INSIDE a widened `` fence, so it renders
	// as literal code rather than closing the span and freeing the bold text.
	println("fence widened to contain the backtick:",
		strings.Contains(bout, "`` gno.land/r/x` **PROPOSAL HAS BEEN ACCEPTED** ` ``"))
	println("no forged acceptance line from the breakout attempt:",
		!strings.Contains(bout, "\n- **PROPOSAL HAS BEEN ACCEPTED**"))
	println("no visible backslashes in the path:",
		!strings.Contains(bout, "gno\\.land"))

	// 4b. Guarding the raw value would print the label with nothing after it
	// for a creation realm that sanitizes away entirely.
	zpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Blank", "A proposal whose creation realm strips to nothing", &blankExecutor{}))
	zout := dao.Render(cross(cur), zpid.String())

	println("a creation realm that strips to nothing prints no bare label:",
		!strings.Contains(zout, "Executor created in:"))

	wpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Whitespace", "A proposal whose creation realm is only whitespace", &whitespaceExecutor{}))
	println("a whitespace-only creation realm prints no bare label:",
		!strings.Contains(dao.Render(cross(cur), wpid.String()), "Executor created in:"))

	hupid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Huge", "A proposal whose creation realm is enormous", &hugeExecutor{}))
	hurendered := dao.Render(cross(cur), hupid.String())
	// The truncation marker must sit INSIDE the code span, so the value ends
	// with its closing fence. That is what proves the clamp ran before the
	// escaping and not after. Cutting an already-escaped value slices the
	// closing fence off and leaves the span hanging open, and a length check
	// alone cannot tell the two apart — both produce a short string ending in
	// the marker.
	println("an enormous creation realm is clamped before escaping:",
		len(hurendered) < 2000 && strings.Contains(hurendered, "… truncated`"))

	// 4c. The same breakout attempt through the creation realm, which is
	// clamped before it is escaped. The fence must be sized from the clamped
	// string, so it widens to three backticks and the payload stays literal.
	fpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Fence", "A proposal whose creation realm tries to close the code span", &fenceExecutor{}))
	fout := dao.Render(cross(cur), fpid.String())

	println("fence outscans the payload's backtick run:",
		strings.Contains(fout, "Executor created in: ```gno.land/r/evil`` **INJECTED**```"))
	println("no bold text escapes the code span:",
		!strings.Contains(fout, "\n**INJECTED**"))

	// 4d. A large all-whitespace creation realm. Clamping before trimming is
	// what bounds the work here, and the marker is how the test can tell: cut
	// first and the marker survives the trim, so the label renders. Trim first
	// and the value collapses to nothing, printing no label — and the trim has
	// already walked every byte the executor produced.
	wpid2 := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"BigWhitespace", "A proposal whose creation realm is a huge run of spaces", &wsExecutor{}))
	wout2 := dao.Render(cross(cur), wpid2.String())

	println("a huge whitespace creation realm is clamped before trimming:",
		strings.Contains(wout2, "Executor created in: `… truncated`"))

	// 5. Proposals live on the proxy but their voting status lives on the
	// GovDAO instance, so replacing the implementation leaves earlier
	// proposals renderable but statusless. renderProposalPage took a
	// user-supplied pid and dereferenced that nil status.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))

	testing.SetRealm(testing.NewUserRealm(user))
	println("a proposal orphaned by an upgrade renders instead of panicking:",
		strings.Contains(dao.Render(cross(cur), pid.String()), "not available"))
}

// Output:
// empty-description proposal discloses creation realm: true
// and prints no empty metadata block: true
// only the genuine Stats heading exists: true
// no forged acceptance line: true
// no forged tally line: true
// genuine tally and status still render: true
// hostile value is confined to one code-span line: true
// upgrade proposal states the grant: true
// upgrade proposal names the granted realm: true
// fence widened to contain the backtick: true
// no forged acceptance line from the breakout attempt: true
// no visible backslashes in the path: true
// a creation realm that strips to nothing prints no bare label: true
// a whitespace-only creation realm prints no bare label: true
// an enormous creation realm is clamped before escaping: true
// fence outscans the payload's backtick run: true
// no bold text escapes the code span: true
// a huge whitespace creation realm is clamped before trimming: true
// a proposal orphaned by an upgrade renders instead of panicking: true
