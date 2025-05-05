package params

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/init"
)

var (
	g1user = testutils.TestAddress("g1user")
)

func init() {
	testing.SetRealm(std.NewUserRealm(g1user))
	init.InitWithUsers(g1user)
}

func TestProUnlockTransfer(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := cross(dao.MustCreateProposal)(pr)
	p, err := cross(dao.GetProposal)(id)
	urequire.NoError(t, err)
	urequire.Equal(t, unlockTransferTitle, p.Title())
}

func TestFailUnlockTransfer(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := cross(dao.MustCreateProposal)(pr)
	urequire.AbortsWithMessage(
		t,
		"proposal didn't reach supermajority yet: 66",
		func() {
			cross(dao.ExecuteProposal)(id)
		},
	)
}

func TestExeUnlockTransfer(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(g1user))

	pr := ProposeUnlockTransferRequest()
	id := cross(dao.MustCreateProposal)(pr)
	p, err := cross(dao.GetProposal)(id)
	urequire.NoError(t, err)
	// urequire.True(t, dao.Active == p.Status()) // TODO

	urequire.NotPanics(
		t,
		func() {
			cross(dao.MustVoteOnProposal)(dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: dao.ProposalID(id),
			})
		},
	)

	urequire.NotPanics(
		t,
		func() {
			cross(dao.ExecuteProposal)(id)
		},
	)
}
