package grants

import (
	"chain"
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
)

const golfPath = "gno.land/r/test/golf"

// Neither is the runner's default caller, so a check that silently runs as
// the default fails instead of passing by coincidence. A call that aborts
// resets the caller: tests set it again before every call.
var (
	member   = testutils.TestAddress("member")
	outsider = testutils.TestAddress("outsider")
	// returnTo is where tests send funds back: the GovDAO treasury realm.
	returnTo = chain.PackageAddress("gno.land/r/gov/dao/treasury/v0")
)

func tranches() []Tranche {
	return []Tranche{{Deliverable: "start", Amount: 80}, {Deliverable: "v3 live", Amount: 320}}
}

// seatMember installs a fresh GovDAO with one T1 member. It takes cur
// rather than setting a realm: realm overrides only hold on the frame that
// sets them.
func seatMember(cur realm) {
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, member, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))
}

func coins(n int64) chain.Coins { return chain.NewCoins(chain.NewCoin(Denom, n)) }

// Grantees are realms. MsgRun scripts are refused too, on a real node:
// gno.land/pkg/integration/testdata/govdao_grants.txtar.
func TestApplyRefusesUsers(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(outsider))
	uassert.AbortsContains(t, cur, "grantees are realms", func() {
		Apply(cross(cur), "t", "p", tranches(), 30*day)
	})
}

func TestOnlyTheGranteeReports(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(golfPath))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/other"))
	uassert.AbortsContains(t, cur, "only the grantee", func() {
		Report(cross(cur), id, "not mine", "")
	})
}

func TestProposalBuildersRefuseBadInput(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(golfPath))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)

	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "no such grant", func() {
		NewApproveRequest(cross(cur), 9999, "Reviewed: test grantee.")
	})
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "reason", func() {
		NewCancelRequest(cross(cur), id, " ")
	})
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "amount must be", func() {
		NewReturnRequest(cross(cur), returnTo, Available()+1, "too much")
	})
}

// Hold is the one call a single member can make against a grant, so every
// way to reach it without being the author of a fresh cancellation of that
// grant is refused.
func TestHoldNeedsTheAuthorOfTheNewestCancellation(cur realm, t *testing.T) {
	seatMember(cur)
	testing.IssueCoins(Address(), coins(Committed()+400))
	testing.SetRealm(testing.NewCodeRealm(golfPath))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)

	testing.SetRealm(testing.NewUserRealm(member))
	approve := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), id, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(approve), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), approve))

	// An approval is not a cancellation.
	uassert.AbortsContains(t, cur, "not the cancellation", func() {
		Hold(cross(cur), id, int64(approve))
	})

	testing.SetRealm(testing.NewUserRealm(member))
	cancel := dao.MustCreateProposal(cross(cur), NewCancelRequest(cross(cur), id, "checking"))

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/other"))
	uassert.AbortsContains(t, cur, "only the author", func() {
		Hold(cross(cur), id, int64(cancel))
	})
	testing.SetRealm(testing.NewUserRealm(outsider))
	uassert.AbortsContains(t, cur, "only the author", func() {
		Hold(cross(cur), id, int64(cancel))
	})

	testing.SetRealm(testing.NewUserRealm(member))
	uassert.NotAborts(t, cur, func() { Hold(cross(cur), id, int64(cancel)) })
	g, _ := desk.Get(id)
	uassert.True(t, g.Held())
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "already used", func() {
		Hold(cross(cur), id, int64(cancel))
	})

	// A fresh cancellation cannot hold the same payment twice.
	testing.SetRealm(testing.NewUserRealm(member))
	again := dao.MustCreateProposal(cross(cur), NewCancelRequest(cross(cur), id, "again"))
	uassert.AbortsContains(t, cur, "already held", func() {
		Hold(cross(cur), id, int64(again))
	})

	// Nor can a cancellation be used once anything else is filed.
	testing.SetRealm(testing.NewUserRealm(member))
	dao.MustCreateProposal(cross(cur), NewCancelRequest(cross(cur), id, "and again"))
	uassert.AbortsContains(t, cur, "not the newest", func() {
		Hold(cross(cur), id, int64(again))
	})
}

// A hold needs the cancellation built in the same block: a cancellation
// GovDAO already voted down, still the newest proposal, cannot hold.
func TestHoldNeedsTheSameBlock(cur realm, t *testing.T) {
	seatMember(cur)
	testing.IssueCoins(Address(), coins(Committed()+400))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/lostvote"))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)
	testing.SetRealm(testing.NewUserRealm(member))
	approve := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), id, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(approve), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), approve))

	testing.SetRealm(testing.NewUserRealm(member))
	cancel := dao.MustCreateProposal(cross(cur), NewCancelRequest(cross(cur), id, "checking"))
	dao.MustVoteOnProposalSimple(cross(cur), int64(cancel), "NO")
	uassert.False(t, dao.ExecuteProposal(cross(cur), cancel))
	testing.SkipHeights(1)
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "same MsgRun", func() { Hold(cross(cur), id, int64(cancel)) })
	// Nor after building a throwaway cancellation: the hold must cite that very build.
	testing.SetRealm(testing.NewUserRealm(member))
	_ = NewCancelRequest(cross(cur), id, "throwaway")
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "same MsgRun", func() { Hold(cross(cur), id, int64(cancel)) })
	uassert.False(t, mustGet(id).Held())
}

// A cancellation filed against a request cannot cancel the grant GovDAO
// approved since: voters were told nothing was reserved.
func TestStaleCancellationFails(cur realm, t *testing.T) {
	seatMember(cur)
	testing.IssueCoins(Address(), coins(Committed()+400))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/stale"))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)
	testing.SetRealm(testing.NewUserRealm(member))
	cancel := dao.MustCreateProposal(cross(cur), NewCancelRequest(cross(cur), id, "Refuse it."))
	approve := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), id, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(cancel), "YES")
	dao.MustVoteOnProposalSimple(cross(cur), int64(approve), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), approve))
	uassert.False(t, dao.ExecuteOrRejectProposal(cross(cur), cancel))
	uassert.Equal(t, "active", mustGet(id).Status.String())
}

// Voters read every deliverable in full: nothing past a cut can hide terms.
func TestApprovalShowsWholeDeliverables(cur realm, t *testing.T) {
	seatMember(cur)
	long := strings.Repeat("strict terms ", 18) + "ESCAPECLAUSE"
	id, err := desk.Apply("gno.land/r/test/longterms", "t", "p", []Tranche{{Deliverable: long, Amount: 80}, {Deliverable: "b", Amount: 320}}, 30*day, time.Now())
	urequire.NoError(t, err)
	testing.SetRealm(testing.NewUserRealm(member))
	pid := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), id, "Reviewed: test grantee."))
	p, _ := dao.GetProposal(pid)
	uassert.True(t, strings.Contains(p.Description(), "ESCAPECLAUSE"))
}

// A namespace counts its own realms' grants, never the wallet accounts the
// gov namespace hosts for everyone.
func TestNamespaceSkipsHostedAccounts(t *testing.T) {
	for _, g := range []string{walletPath + "#team1", walletPath + "#team2", walletPath + "#team3", "gno.land/r/gov/x"} {
		_, err := desk.Apply(g, "t", "p", tranches(), 30*day, time.Now())
		urequire.NoError(t, err)
	}
	list, n := nsGrants("gov", 2)
	uassert.Equal(t, 1, n)
	urequire.Equal(t, 1, len(list))
	uassert.Equal(t, "gno.land/r/gov/x", list[0].Grantee)
	_, n = nsGrants("team1", 0)
	uassert.Equal(t, 1, n)
	// A wallet account's namespace may carry a dot; its page must open.
	_, err := desk.Apply(walletPath+"#dot.ns", "t", "p", tranches(), 30*day, time.Now())
	urequire.NoError(t, err)
	uassert.False(t, strings.HasPrefix(Render("ns/dot.ns"), "# Not found"))
}

// A passed approval the pool cannot cover panics instead of failing, so
// nobody can turn a funding gap into a permanent rejection. That it then
// executes once the pool is funded needs a real transaction revert, so it is
// tested on a node: govdao_grants.txtar.
func TestUnderfundedApprovalPanics(cur realm, t *testing.T) {
	seatMember(cur)
	// Coins do not carry over between test functions; realm state does.
	testing.IssueCoins(Address(), coins(Committed()))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/poor"))
	id := Apply(cross(cur), "t", "p", []Tranche{{Deliverable: "all", Amount: 1_000_000}}, 0)

	testing.SetRealm(testing.NewUserRealm(member))
	pid := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), id, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	uassert.AbortsContains(t, cur, "execute again", func() {
		dao.ExecuteOrRejectProposal(cross(cur), pid)
	})
	g, _ := desk.Get(id)
	uassert.Equal(t, "pending", g.Status.String())
}

// Grantee words are quoted and escaped: they cannot draw the desk's own
// headings, tables or scripts.
func TestRenderEscapesGranteeText(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/inject"))
	id := Apply(cross(cur), "[click](javascript:alert(1)) | x", "## For GovDAO members\n\n```go\nevil()\n```",
		[]Tranche{{Deliverable: "a|b", Amount: 1}}, 0)
	page := Render("grant/" + itoa(id))
	uassert.False(t, strings.Contains(page, "[click](javascript"), "title link survived")
	uassert.True(t, strings.Contains(page, quote("## For GovDAO members\n\n```go\nevil()\n```")), "pitch not quoted and escaped")
	uassert.False(t, strings.Contains(page, "```go\nevil"), "pitch opened a code fence")
	uassert.True(t, strings.Contains(page, `a\|b`), "deliverable pipe not escaped")

	for _, p := range []string{"grant/0", "ns/Bad Name", "all/x", "all/0", "all/99", "active/99999999999", "waiting/0", "grant/1/x"} {
		uassert.True(t, strings.HasPrefix(Render(p), "# Not found"), p)
	}
	uassert.True(t, strings.Contains(Render("ns/test"), "r/test/inject"))
}

func TestGnot(t *testing.T) {
	uassert.Equal(t, "0 GNOT", gnot(0))
	uassert.Equal(t, "1.5 GNOT", gnot(1_500_000))
	uassert.Equal(t, "0.000001 GNOT", gnot(1))
	uassert.Equal(t, "30,000 GNOT", gnot(30_000_000_000))
	uassert.Equal(t, "1,234,567.89 GNOT", gnot(1_234_567_890_000))
	uassert.Equal(t, "-0.000001 GNOT", gnot(-1))
	uassert.Equal(t, "-1.5 GNOT", gnot(-1_500_000))
	uassert.Equal(t, "-123 GNOT", gnot(-123_000_000))
	uassert.Equal(t, "-9,223,372,036,854.775808 GNOT", gnot(-1<<63))
}
