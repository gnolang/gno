package params

import (
	"testing"

	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
)

func TestSetFeeCollector(t *testing.T) {
	userRealm := testing.NewUserRealm(g1user)
	testing.SetRealm(userRealm)

	pr := NewSetFeeCollectorRequest(userRealm.Address())
	id := dao.MustCreateProposal(cross, pr)
	_, err := dao.GetProposal(cross, id)
	urequire.NoError(t, err)

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

	// XXX: test that the value got properly updated, when we can get params from gno code
}
