package params

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/urequire"
	"gno.land/r/gov/dao"
	ini "gno.land/r/gov/dao/v3/init"
)

var g1user = testutils.TestAddress("g1user")

func init() {
	testing.SetRealm(testing.NewUserRealm(g1user))
	ini.InitWithUsers(g1user)
}

func TestProUnlockTransfer(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(cross, pr)
	p, err := dao.GetProposal(cross, id)
	urequire.NoError(t, err)
	urequire.Equal(t, unlockTransferTitle, p.Title())
}

func TestFailUnlockTransfer(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(cross, pr)
	urequire.AbortsWithMessage(
		t,
		"proposal didn't reach supermajority yet: 66.66",
		func() {
			dao.ExecuteProposal(cross, id)
		},
	)
}

func TestExeUnlockTransfer(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(cross, pr)
	_, err := dao.GetProposal(cross, id)
	urequire.NoError(t, err)
	// urequire.True(t, dao.Active == p.Status()) // TODO

	urequire.NotPanics(
		t,
		func() {
			dao.MustVoteOnProposal(cross, dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: dao.ProposalID(id),
			})
		},
	)

	urequire.NotPanics(
		t,
		func() {
			dao.ExecuteProposal(cross, id)
		},
	)
}
