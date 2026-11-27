package grants

import (
	"chain/banker"
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/r/gov/dao"
)

// Named zz_* to run last: closing the desk is final for every later test.
//
// Close cancels the active grants, returns the whole pool, and a second
// Close passed meanwhile fails with an error, so it can be rejected.
func TestCloseReturnsEverythingOnce(cur realm, t *testing.T) {
	seatMember(cur)
	testing.IssueCoins(Address(), coins(Committed()+500))
	treasury := returnTo
	before := banker.NewReadonlyBanker().GetCoin(treasury, Denom)
	pool := Balance()
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/late/waiter"))
	waiter := Apply(cross(cur), "t", "p", tranches(), 30*day)

	testing.SetRealm(testing.NewUserRealm(member))
	first := dao.MustCreateProposal(cross(cur), NewCloseRequest(cross(cur), returnTo, "end of round", true))
	second := dao.MustCreateProposal(cross(cur), NewCloseRequest(cross(cur), returnTo, "again", true))
	dao.MustVoteOnProposalSimple(cross(cur), int64(first), "YES")
	dao.MustVoteOnProposalSimple(cross(cur), int64(second), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), first))

	uassert.Equal(t, int64(0), Committed())
	uassert.Equal(t, int64(0), Balance())
	uassert.Equal(t, before+pool, banker.NewReadonlyBanker().GetCoin(treasury, Denom))
	uassert.Equal(t, 0, desk.ActiveCount())
	uassert.False(t, dao.ExecuteOrRejectProposal(cross(cur), second))

	// A request left pending is no longer offered to GovDAO.
	uassert.AbortsContains(t, cur, "closed", func() { NewApproveRequest(cross(cur), waiter, "Reviewed: test grantee.") })
	page := Render("grant/" + itoa(waiter))
	uassert.True(t, strings.Contains(page, "Desk closed"))
	uassert.False(t, strings.Contains(page, "NewApproveRequest"), "no sponsor script once closed")
	uassert.True(t, strings.Contains(Render(""), "desk closed"))

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/late/comer"))
	uassert.AbortsContains(t, cur, "closed", func() {
		Apply(cross(cur), "t", "p", tranches(), 30*day)
	})
}
