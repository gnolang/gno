package grants

import (
	"strings"
	"testing"
	"time"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

// A frozen grantee is paid nothing, cannot apply or be approved, and its
// pages say so without offering a Claim that would abort. Unfrozen, its due
// payment is released as usual.
func TestFreezeAndUnfreeze(cur realm, t *testing.T) {
	const who = "gno.land/r/frozen/team"
	seatMember(cur)
	testing.SetRealm(testing.NewCodeRealm(who))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)
	testing.IssueCoins(Address(), coins(Committed()+400))
	testing.SetRealm(testing.NewUserRealm(member))
	pid := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), id, "Reviewed."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), pid))

	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "ever applied", func() {
		NewFreezeRequest(cross(cur), "gno.land/r/nobody/here", true, "r")
	})
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "already unfrozen", func() { NewFreezeRequest(cross(cur), who, false, "r") })
	testing.SetRealm(testing.NewUserRealm(member))
	freeze := dao.MustCreateProposal(cross(cur), NewFreezeRequest(cross(cur), who, true, "Spends unexplained."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(freeze), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), freeze))
	uassert.True(t, Frozen(who))
	uassert.Equal(t, "Spends unexplained.", FrozenReason(who))

	start := time.Now()
	setNow(start.Add(14 * day))
	testing.SetRealm(testing.NewUserRealm(outsider))
	uassert.AbortsContains(t, cur, "GovDAO froze", func() { Claim(cross(cur), id) })
	testing.SetRealm(testing.NewCodeRealm(who))
	uassert.AbortsContains(t, cur, "GovDAO froze", func() { Apply(cross(cur), "t", "p", tranches(), 30*day) })
	page := Render("grant/" + itoa(id))
	uassert.True(t, strings.Contains(page, "is due and waits"))
	uassert.Equal(t, 0, strings.Count(page, `<gno-form exec="Claim">`))
	uassert.True(t, strings.Contains(page, "Unfreeze the grantee"))
	uassert.False(t, strings.Contains(Render(""), "#"+itoa(id)+" t](/r/gov/grants/v0:grant/"+itoa(id)+"): payment 1 can be released"))

	testing.SetRealm(testing.NewUserRealm(member))
	unfreeze := dao.MustCreateProposal(cross(cur), NewFreezeRequest(cross(cur), who, false, "Explained."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(unfreeze), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), unfreeze))
	testing.SetRealm(testing.NewUserRealm(outsider))
	uassert.NotAborts(t, cur, func() { Claim(cross(cur), id) })
	setNow(start)
}

// A grantee frozen after its approval was proposed: execution fails with an
// error, so the vote can be rejected.
func TestFreezeAfterProposalMakesApprovalRejectable(cur realm, t *testing.T) {
	const who = "gno.land/r/frozen/late"
	seatMember(cur)
	testing.SetRealm(testing.NewCodeRealm(who))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)
	testing.IssueCoins(Address(), coins(Committed()+400))
	testing.SetRealm(testing.NewUserRealm(member))
	approve := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), id, "Reviewed."))
	freeze := dao.MustCreateProposal(cross(cur), NewFreezeRequest(cross(cur), who, true, "Doubts."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(approve), "YES")
	dao.MustVoteOnProposalSimple(cross(cur), int64(freeze), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), freeze))
	uassert.False(t, dao.ExecuteOrRejectProposal(cross(cur), approve))
	g, _ := desk.Get(id)
	uassert.Equal(t, "pending", g.Status.String())
}

// The sponsor builds the approval in their own MsgRun, with a review the
// approval will accept.
func TestApprovalReviewIsCheckedUpFront(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/review/me"))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "lines", func() {
		NewApproveRequest(cross(cur), id, strings.Repeat("ok\n", 50))
	})
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/some/realm"))
	uassert.AbortsContains(t, cur, "own MsgRun", func() { NewApproveRequest(cross(cur), id, "Reviewed.") })
}

// A wallet account's grants count for its namespace, never for gov, where
// the wallet lives; the approval text tells voters so.
func TestWalletGrantsCountForTheirNamespace(cur realm, t *testing.T) {
	grantee := walletPath + "#walletteam"
	id, err := desk.Apply(grantee, "Wallet team grant", "p", tranches(), 30*day, time.Now())
	urequire.NoError(t, err)
	testing.SetRealm(testing.NewUserRealm(member))
	text := approveText(mustGet(id), member.String(), "Reviewed.")
	uassert.True(t, strings.Contains(text, "this one included: 1"), text)
	uassert.True(t, strings.Contains(text, "check its owner there"))
	uassert.True(t, strings.Contains(Render("ns/walletteam"), "Wallet team grant"))
	uassert.False(t, strings.Contains(Render("ns/gov"), "Wallet team grant"))
	uassert.Equal(t, "walletteam", namespace(grantee))
}

// Labels name what a grantee did, as every page shows it.
func TestLabels(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/label/quit"))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/label/quit"))
	Withdraw(cross(cur), id)
	g, _ := desk.Get(id)
	uassert.Equal(t, "withdrawn", Label(g))
	uassert.True(t, strings.Contains(Render("grant/"+itoa(id)), "[!CAUTION] Withdrawn"))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/label/other"))
	uassert.AbortsContains(t, cur, "only the grantee", func() { Withdraw(cross(cur), id) })
}

// A hold must cite the desk's own cancellation: a proposal another realm's
// builder titled the same way (r/sys/params takes any title) is refused.
func TestHoldRefusesALookAlikeCancellation(cur realm, t *testing.T) {
	seatMember(cur)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/hold/forged"))
	id := Apply(cross(cur), "Forged", "p", tranches(), 30*day)
	testing.IssueCoins(Address(), coins(Committed()+400))
	testing.SetRealm(testing.NewUserRealm(member))
	pid := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), id, "Reviewed."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), pid))

	testing.SetRealm(testing.NewUserRealm(member))
	fake := dao.MustCreateProposal(cross(cur), params.NewSysParamStringsPropRequestWithTitle(cross(cur),
		"bank", "p", "restricted_denoms", cancelTitle(mustGet(id)), []string{"ugnot"}))
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "is not the cancellation", func() { Hold(cross(cur), id, int64(fake)) })
}

// A frozen realm cannot bring new money in under a fresh sub-identity, even
// one that only ever applied through identities it hosts. Identities already
// approved stay payable, and their pages say so.
func TestFrozenHostCannotBringNewIdentities(cur realm, t *testing.T) {
	const host = "gno.land/r/frozen/host"
	seatMember(cur)
	a, err := desk.Apply(host+"#a", "A", "p", tranches(), 30*day, time.Now())
	urequire.NoError(t, err)
	urequire.NoError(t, desk.Approve(a, 1<<50, "", "", time.Now()))
	testing.IssueCoins(Address(), coins(Committed()))
	testing.SetRealm(testing.NewUserRealm(member))
	freeze := dao.MustCreateProposal(cross(cur), NewFreezeRequest(cross(cur), host, true, "Diverts funds."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(freeze), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), freeze))

	id, err := desk.Apply(host+"#fresh", "Fresh", "p", tranches(), 30*day, time.Now())
	urequire.NoError(t, err)
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "or its host", func() { NewApproveRequest(cross(cur), id, "Reviewed.") })
	page := Render("grant/" + itoa(id))
	uassert.True(t, strings.Contains(page, "Grantee frozen"))
	uassert.False(t, strings.Contains(page, "NewApproveRequest(cross(cur), "+itoa(id)), "no approval script")
	uassert.True(t, strings.Contains(nextStep(mustGet(id)), "grantee frozen by GovDAO"))
	uassert.True(t, strings.Contains(badge(mustGet(id)), "frozen"))
	uassert.False(t, strings.Contains(badge(mustGet(a)), "frozen"), "an approved identity is still paid")
}

// Placeholders copied from the page, the reason reserved for withdrawals and
// a return to the pool itself are refused when the proposal is built.
func TestProposalInputsAreChecked(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/inputs/team"))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "write your own review", func() {
		NewApproveRequest(cross(cur), id, "REVIEW: who controls the grantee")
	})
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "write a reason", func() { NewCancelRequest(cross(cur), id, "REASON") })
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "reserved", func() { NewCancelRequest(cross(cur), id, "withdrawn by the grantee") })
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "other than the pool", func() { NewReturnRequest(cross(cur), Address(), 1, "r") })
	testing.SetRealm(testing.NewUserRealm(member))
	uassert.AbortsContains(t, cur, "valid address", func() { NewCloseRequest(cross(cur), "nope", "r", false) })
}
