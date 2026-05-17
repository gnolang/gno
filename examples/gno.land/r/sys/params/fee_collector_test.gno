package params

import (
	"testing"

	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
func TestSetFeeCollector(cur realm, t *testing.T) {
	userRealm := testing.NewUserRealm(g1user)
	testing.SetRealm(userRealm)

	pr := NewSetFeeCollectorRequest(cur, userRealm.Address())
	id := dao.MustCreateProposal(cross2(cur), pr)
	_, err := dao.GetProposal(id)
	urequire.NoError(t, err)

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

	// XXX: test that the value got properly updated, when we can get params from gno code
}
