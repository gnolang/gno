package params

import (
	"std"
	"testing"

	"gno.land/p/demo/urequire"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/init"
)

func init() {
	c := std.OriginCaller()
	init.InitWithUsers(c)
}

func TestProUnlockTransfer(t *testing.T) {
	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(pr)
	p, err := dao.GetProposal(id)
	urequire.NoError(t, err)
	urequire.Equal(t, unlockTransferTitle, p.Title())
}

func TestFailUnlockTransfer(t *testing.T) {
	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(pr)
	urequire.AbortsWithMessage(
		t,
		"proposal didn't reach supermajority yet: 66",
		func() {
			dao.ExecuteProposal(id)
		},
	)
}

func TestExeUnlockTransfer(t *testing.T) {
	pr := ProposeUnlockTransferRequest()
	id := dao.MustCreateProposal(pr)
	p, err := dao.GetProposal(id)
	urequire.NoError(t, err)
	// urequire.True(t, dao.Active == p.Status()) // TODO

	urequire.NotAborts(
		t,
		func() {
			dao.MustVoteOnProposal(dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: dao.ProposalID(id),
			})
		},
	)

	urequire.NotAborts(
		t,
		func() {
			dao.ExecuteProposal(id)
		},
	)
}
