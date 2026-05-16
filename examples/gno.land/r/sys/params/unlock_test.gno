package params

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
	ini "gno.land/r/gov/dao/v3/init"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
var g1user = testutils.TestAddress("g1user")

func init() {
	testing.SetRealm(testing.NewUserRealm(g1user))
	ini.InitWithUsers(g1user)
}

func TestProUnlockTransfer(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(cross2(cur), pr)
	p, err := dao.GetProposal(cross2(cur), id)
	urequire.NoError(t, err)
	urequire.Equal(t, unlockTransferTitle, p.Title())
}

func TestFailUnlockTransfer(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(cross2(cur), pr)
	urequire.AbortsWithMessage(
		t, cur,
		"proposal didn't reach supermajority yet: 66.66",
		func() {
			dao.ExecuteProposal(cross2(cur), id)
		},
	)
}

func TestExeUnlockTransfer(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(cross2(cur), pr)
	_, err := dao.GetProposal(cross2(cur), id)
	urequire.NoError(t, err)
	// urequire.True(t, dao.Active == p.Status()) // TODO

	urequire.NotPanics(
		t, cur,
		func() {
			dao.MustVoteOnProposal(cross2(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(id)))
		},
	)

	urequire.NotPanics(
		t, cur,
		func() {
			dao.ExecuteProposal(cross2(cur), id)
		},
	)
}
