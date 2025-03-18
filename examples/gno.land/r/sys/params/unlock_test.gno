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
	id := ProposeUnlockTransfer()
	p, err := dao.GetProposal(id)
	urequire.NoError(t, err)
	urequire.Equal(t, unlockTransferTitle, p.Title())
}

func TestFailUnlockTransfer(t *testing.T) {
	id := ProposeUnlockTransfer()
	urequire.PanicsWithMessage(
		t,
		"TODO",
		func() {
			dao.ExecuteProposal(id)
		},
	)
}

func TestExeUnlockTransfer(t *testing.T) {
	id := ProposeUnlockTransfer()
	p, err := dao.GetProposal(id)
	urequire.NoError(t, err)
	// urequire.True(t, dao.Active == p.Status()) // TODO

	urequire.NotPanics(
		t,
		func() {
			dao.MustVoteOnProposal(dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: dao.ProposalID(id),
			})
		},
	)

	urequire.NotPanics(
		t,
		func() {
			dao.ExecuteProposal(id)
		},
	)
}
