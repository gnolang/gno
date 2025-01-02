package params

import (
	"testing"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/simpledao"
	"gno.land/p/demo/urequire"
	"gno.land/r/gov/dao/bridge"
)

func TestProUnlockSend(t *testing.T) {
	govdao := bridge.GovDAO()
	id := ProposeUnlockSend()
	p, err := govdao.GetPropStore().ProposalByID(id)
	urequire.NoError(t, err)
	urequire.Equal(t, UnlockSendTitle, p.Title())
}

func TestFailUnlockSend(t *testing.T) {
	govdao := bridge.GovDAO()
	id := ProposeUnlockSend()
	urequire.PanicsWithMessage(
		t,
		simpledao.ErrProposalNotAccepted.Error(),
		func() {
			govdao.ExecuteProposal(id)
		},
	)
}

func TestExeUnlockSend(t *testing.T) {
	govdao := bridge.GovDAO()
	id := ProposeUnlockSend()
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
