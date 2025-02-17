package params

import (
	"testing"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/simpledao"
	"gno.land/p/demo/urequire"
	"gno.land/r/gov/dao/bridge"
)

func TestProUnlockTransfer(t *testing.T) {
	govdao := bridge.GovDAO()
	id := ProposeUnlockTransfer()
	p, err := govdao.GetPropStore().ProposalByID(id)
	urequire.NoError(t, err)
	urequire.Equal(t, UnlockTransferTitle, p.Title())
}

func TestFailUnlockTransfer(t *testing.T) {
	govdao := bridge.GovDAO()
	id := ProposeUnlockTransfer()
	urequire.PanicsWithMessage(
		t,
		simpledao.ErrProposalNotAccepted.Error(),
		func() {
			govdao.ExecuteProposal(id)
		},
	)
}

func TestExeUnlockTransfer(t *testing.T) {
	govdao := bridge.GovDAO()
	id := ProposeUnlockTransfer()
	p, err := govdao.GetPropStore().ProposalByID(id)
	urequire.NoError(t, err)
	urequire.True(t, dao.Active == p.Status())

	govdao.VoteOnProposal(id, dao.YesVote)

	urequire.True(t, dao.Accepted == p.Status())

	urequire.NotPanics(
		t,
		func() {
			govdao.ExecuteProposal(id)
		},
	)

	urequire.True(t, dao.ExecutionSuccessful == p.Status())
}
