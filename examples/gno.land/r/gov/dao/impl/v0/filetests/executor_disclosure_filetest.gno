// PKGPATH: gno.land/r/test/disclosure
package disclosure

// Covers the executor-disclosure changes in an isolated realm. The unit tests
// in the impl package share proposal ids across files (govdao_test.gno asserts
// a hard-coded id) and swap the DAO implementation partway through, so these
// live here instead.

import (
	"strconv"
	"strings"
	"testing"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
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

// pathExecutor reports whatever creation realm it is handed, so one type covers
// every shape the source link must refuse to build.
type pathExecutor struct{ path string }

func (e *pathExecutor) Execute(cur realm) error { return nil }

func (e *pathExecutor) String() string { return "" }

func (e *pathExecutor) CreationRealm() string { return e.path }

// quotedBody reports whether every line between the page's Description
// heading and its closing rule carries the quote marker, which is what keeps
// whatever the proposer wrote from rendering outside the quote.
func quotedBody(page string) bool {
	return quotedSection(page, "### Description\n\n", "\n\n---\n\n### Actions")
}

// quotedSection is quotedBody for any section: every line between the heading
// and the page's own text that closes it carries the quote marker.
func quotedSection(page, heading, closing string) bool {
	start := strings.Index(page, heading)
	end := strings.LastIndex(page, closing)
	if start < 0 || end < start {
		return false
	}

	for _, line := range strings.Split(page[start+len(heading):end], "\n") {
		if !strings.HasPrefix(line, "  >") {
			return false
		}
	}

	return true
}

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
		strings.Contains(out, "Executes: [`gno.land/r/template/silent`](/r/template/silent$source)"))
	// An executor with nothing to say renders no Execution section at all: the
	// realm is in the metadata block above, and there is nothing else to show.
	println("and an executor with nothing to say gets no section:",
		!strings.Contains(out, "### Execution"))
	// The DAO's section comes first, and the page's own rule separates it from
	// the description under it. Nothing the proposer wrote precedes it.
	println("and the description sits below the rule:",
		strings.Contains(out, "---\n\n### Description\n\n  > A proposal whose executor has no description"))

	// 2. CreationRealm() is dispatched through the public dao.Executor
	// interface, so a third-party executor picks its own value — and the
	// disclosure now renders for every proposal. InlineCode does not delete the
	// hostile text; it folds it onto one line inside a code span, where it can
	// no longer forge page structure. So assert structure, not absence.
	hpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Hostile", "A proposal whose executor forges page structure", &hostileExecutor{}))
	hout := dao.Render(cross(cur), hpid.String())

	println("the payload forges no heading:",
		!strings.Contains(hout, "\n### Stats"))
	println("no forged acceptance line:",
		!strings.Contains(hout, "\n- **PROPOSAL HAS BEEN ACCEPTED**"))
	println("no forged tally line:",
		!strings.Contains(hout, "\n- YES PERCENT: 100%"))
	println("genuine tally and status still render:",
		strings.Contains(hout, "Votes: YES 0.00%") &&
			strings.Contains(hout, "Status: ACTIVE"))
	// One line, inside a code span: the newlines are folded to spaces, so none
	// of the payload can start a block. (No pad space before the value now —
	// TrimSpace removes the payload's trailing newline, so InlineCode does not
	// need to pad the fence.)
	println("hostile value is confined to one code-span line:",
		strings.Contains(hout, "Executes: `gno.land/r/sys/params  ### Stats"))
	// It opens with a real realm path and is not one: the value carries the rest
	// of the payload, so the grammar check rejects it and no link is built.
	println("and carries no link:", !strings.Contains(hout, "$source"))

	// 3. The upgrade proposal rewrites AllowedDAOs — the sole authorization for
	// replacing the implementation, mutating the member store and moving
	// treasury funds. It used to carry an empty description, so the realm
	// receiving that authority appeared nowhere a voter would read.
	upid := dao.MustCreateProposal(cross(cur),
		impl.NewUpgradeDaoImplRequest(cross(cur), impl.NewGovDAO(), "gno.land/r/gov/dao/v1/impl", "reason"))
	uout := dao.Render(cross(cur), upid.String())

	println("upgrade proposal states the grant:",
		strings.Contains(uout, "may replace the implementation, mutate the member store, or move treasury funds"))
	println("upgrade proposal names the granted realm:",
		strings.Contains(uout, "`gno.land/r/gov/dao/v1/impl`"))

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

	println("a creation realm that strips to nothing says so:",
		strings.Contains(zout, "Executes: none reported"))

	wpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Whitespace", "A proposal whose creation realm is only whitespace", &whitespaceExecutor{}))
	println("a whitespace-only creation realm says so too:",
		strings.Contains(dao.Render(cross(cur), wpid.String()), "Executes: none reported"))

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
		strings.Contains(fout, "Executes: ```gno.land/r/evil`` **INJECTED**```"))
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
		strings.Contains(wout2, "Executes: `… truncated`"))

	// 4e. A pure package is linked too: an executor created in one reports that
	// path, and gnoweb serves its source the same way.
	ppid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Pure", "A proposal whose executor was created in a pure package",
		&pathExecutor{path: "gno.land/p/nt/ufmt/v0"}))
	println("a pure package links to its source too:",
		strings.Contains(dao.Render(cross(cur), strconv.FormatInt(int64(ppid), 10)),
			"Executes: [`gno.land/p/nt/ufmt/v0`](/p/nt/ufmt/v0$source)"))

	// 4f. The link is built only for a package path this chain serves, so a
	// proposal cannot aim it. Everything else keeps its code span.
	for _, path := range []string{
		"example.com/r/foo", // another chain
		"gno.land/e/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/run", // an ephemeral realm, which has no source page
		"gno.land/r/Evil",                     // uppercase, which no deployed path holds
		"gno.land/r/a--b",                     // two separators in a row
		"gno.land/r/",                         // nothing after the letter
		"gno.land/r/x)](javascript:alert(1))", // a destination of the executor's choosing
	} {
		lpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
			"Path", "A proposal whose creation realm is not a package path", &pathExecutor{path: path}))
		// Decimal, not ProposalID.String(): the router parses the page's own
		// links, and a seqid past 9 carries letters that ParseInt rejects.
		lout := dao.Render(cross(cur), strconv.FormatInt(int64(lpid), 10))

		// The last case writes a link destination of its own, which stays
		// literal text between the fences.
		println("unlinked, and confined to its code span:", path,
			!strings.Contains(lout, "$source") &&
				strings.Contains(lout, "Executes: `"+path+"`"))
	}

	// 4g. Everything the proposer writes renders inside a quote the page draws,
	// so no block a description opens can reach the page's rule, the
	// Description heading or the action bar below it. One case per shape that
	// got past an escaping pass, each forging an Execution section where it can
	// carry one.
	hostile := []struct{ name, description string }{
		{"a plain rule", "Looks fine.\n\n---\n\n### Execution\n\nExecutor created in: `gno.land/r/safe/thing`"},
		{"a CRLF rule", "Looks fine.\r\n\r\n---\r\n\r\n### Execution\r\n"},
		{"a carriage return inside the run", "Looks fine.\n\n--\r-\n\n### Execution\n"},
		{"a carriage return inside a fence", "``\r`\n\n---\n\n### Execution\n"},
		{"a backtick in an info string", "Looks fine.\n\n```a`b\n\n---\n\n### Execution\n"},
		{"a backtick fence inside a tilde block", "~~~\n```\n~~~\n\n---\n\n### Execution\n"},
		{"an open fence", "Looks fine.\n\n```\nstill code"},
		{"an open HTML block", "Looks fine.\n\n<pre>\nstill raw"},
		{"an open comment", "Looks fine.\n\n<!--\nstill hidden"},
		{"a stray angle bracket before a fence", "<3 hearts\n```\n\n---\n\n### Execution\n"},
		{"a rule indented under a list item", "- item\n\n    ---\n\n### Execution\n"},
		{"a list continuation over dashes", "- item\ncontinued\n---\n\n### Execution\n"},
		{"spaced dashes under a paragraph", "My Section\n- - -\n\n### Execution\n"},
		{"a lazy line after a quote", "> q\nlazy\n---\n\n### Execution\n"},
	}
	for _, h := range hostile {
		qpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
			"Hostile", h.description, &pathExecutor{path: "gno.land/r/template/silent"}))
		qout := dao.Render(cross(cur), strconv.FormatInt(int64(qpid), 10))

		println("contained:", h.name, quotedBody(qout) && strings.Count(qout, "\n---\n") == 2)
	}

	// 4h. The forged section stays below the real one, and honest markdown is
	// left exactly as the proposer wrote it, inside the quote.
	rpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Rule",
		"My Section\n---\n\nLooks fine.\n\n### Execution\n\nExecutor created in: `gno.land/r/safe/thing`",
		&pathExecutor{path: "gno.land/r/template/silent"}))
	rout := dao.Render(cross(cur), strconv.FormatInt(int64(rpid), 10))

	println("the real disclosure is above the forged one:",
		strings.Index(rout, "Executes:") < strings.Index(rout, "### Description"))
	println("an honest heading is left as written:",
		strings.Contains(rout, "  > My Section\n  > ---\n"))

	// 4i. The quote rewrites every line and the render pays for it, so the
	// description is clamped first like every other slot on the page.
	opid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Enormous description",
		strings.Repeat("a-\n", 60000),
		&pathExecutor{path: "gno.land/r/template/silent"}))
	oout := dao.Render(cross(cur), strconv.FormatInt(int64(opid), 10))

	println("an enormous description is clamped:",
		strings.Contains(oout, "… truncated") && len(oout) < 20000)

	// 4j. A footnote is the one block goldmark moves out of its container, to
	// the foot of the page below the action bar, so no footnote may form.
	npid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Footnote",
		"Looks fine.[^1]\n\n[^1]: ### Execution\n\n    Executor created in: `gno.land/r/safe/thing`",
		&pathExecutor{path: "gno.land/r/template/silent"}))
	nout := dao.Render(cross(cur), strconv.FormatInt(int64(npid), 10))

	println("a footnote definition stays inside the quote:",
		quotedBody(nout) && strings.Contains(nout, "  > \\[^1]: ### Execution"))

	// 4k. gnoweb draws a quote whose first line is `[!KIND]` as an alert box, so
	// a description opening on one would style the page's quote as a callout.
	apid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Alert", "[!WARNING]\nVerified by the DAO.",
		&pathExecutor{path: "gno.land/r/template/silent"}))
	aout := dao.Render(cross(cur), strconv.FormatInt(int64(apid), 10))

	println("a leading alert marker stays literal:",
		strings.Contains(aout, "### Description\n\n  > \\[!WARNING]\n  > Verified"))

	// 4m. Honest markdown renders as the proposer wrote it. A `[^` with no `]:`
	// after it on its line opens no footnote and keeps its bytes, which a
	// character class in code depends on, and a tab keeps its width, since the
	// quote starts each line at column 4: tab-indented code stays code.
	hpid2 := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Honest", "Matches `[^a-z]` names:\n\n\tre := regexp.MustCompile(\"[^0-9]\")",
		&pathExecutor{path: "gno.land/r/template/silent"}))
	hout2 := dao.Render(cross(cur), strconv.FormatInt(int64(hpid2), 10))

	println("a caret in code keeps its bytes:",
		strings.Contains(hout2, "  > Matches `[^a-z]` names:") &&
			strings.Contains(hout2, "  > \tre := regexp.MustCompile(\"[^0-9]\")"))

	// 4l. A built-in executor's description carries what the proposer wrote, so
	// the Execution section is quoted like the Description below it.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/template/silent"))
	forger := dao.NewSimpleExecutor(0, cur, func(realm) error { return nil },
		"Pays a grant.\n\n---\n\n### Description\n\nLooks fine.\n\n```\nstill code")
	testing.SetRealm(testing.NewUserRealm(user))
	epid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Executor", "An honest description", forger))
	eout := dao.Render(cross(cur), strconv.FormatInt(int64(epid), 10))

	println("an executor description is contained:",
		quotedSection(eout, "### Execution\n\n", "\n\n---\n\n### Description\n\n  > An honest") &&
			strings.Count(eout, "\n---\n") == 2)

	// 5. Proposals live on the proxy but their voting status lives on the
	// GovDAO instance, so replacing the implementation leaves earlier
	// proposals renderable but statusless. renderProposalPage took a
	// user-supplied pid and dereferenced that nil status.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/impl/v0"))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))

	testing.SetRealm(testing.NewUserRealm(user))
	println("a proposal orphaned by an upgrade renders instead of panicking:",
		strings.Contains(dao.Render(cross(cur), pid.String()), "not available"))
}

// Output:
// empty-description proposal discloses creation realm: true
// and an executor with nothing to say gets no section: true
// and the description sits below the rule: true
// the payload forges no heading: true
// no forged acceptance line: true
// no forged tally line: true
// genuine tally and status still render: true
// hostile value is confined to one code-span line: true
// and carries no link: true
// upgrade proposal states the grant: true
// upgrade proposal names the granted realm: true
// fence widened to contain the backtick: true
// no forged acceptance line from the breakout attempt: true
// no visible backslashes in the path: true
// a creation realm that strips to nothing says so: true
// a whitespace-only creation realm says so too: true
// an enormous creation realm is clamped before escaping: true
// fence outscans the payload's backtick run: true
// no bold text escapes the code span: true
// a huge whitespace creation realm is clamped before trimming: true
// a pure package links to its source too: true
// unlinked, and confined to its code span: example.com/r/foo true
// unlinked, and confined to its code span: gno.land/e/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/run true
// unlinked, and confined to its code span: gno.land/r/Evil true
// unlinked, and confined to its code span: gno.land/r/a--b true
// unlinked, and confined to its code span: gno.land/r/ true
// unlinked, and confined to its code span: gno.land/r/x)](javascript:alert(1)) true
// contained: a plain rule true
// contained: a CRLF rule true
// contained: a carriage return inside the run true
// contained: a carriage return inside a fence true
// contained: a backtick in an info string true
// contained: a backtick fence inside a tilde block true
// contained: an open fence true
// contained: an open HTML block true
// contained: an open comment true
// contained: a stray angle bracket before a fence true
// contained: a rule indented under a list item true
// contained: a list continuation over dashes true
// contained: spaced dashes under a paragraph true
// contained: a lazy line after a quote true
// the real disclosure is above the forged one: true
// an honest heading is left as written: true
// an enormous description is clamped: true
// a footnote definition stays inside the quote: true
// a leading alert marker stays literal: true
// a caret in code keeps its bytes: true
// an executor description is contained: true
// a proposal orphaned by an upgrade renders instead of panicking: true
