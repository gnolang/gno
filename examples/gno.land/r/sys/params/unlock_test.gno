package params

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
	ini "gno.land/r/gov/dao/v3/init"
)

var g1user = testutils.TestAddress("g1user")

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(g1user))
	ini.InitWithUsers(cross2(cur), g1user)
}

func TestProUnlockTransfer(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(cross2(cur), pr)
	p, err := dao.GetProposal(id)
	urequire.NoError(t, err)
	urequire.Equal(t, unlockTransferTitle, p.Title())
}

func TestFailUnlockTransfer(cur realm, t *testing.T) {
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

func TestExeUnlockTransfer(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(cross2(cur), pr)
	_, err := dao.GetProposal(id)
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
