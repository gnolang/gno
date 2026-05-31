package commondao_test

import (
	"testing"

	"gno.land/p/nt/urequire/v0"

	"gno.land/p/nt/commondao/v0"
)

func TestNewVotingContext(t *testing.T) {
	r := &commondao.VotingRecord{}
	s := commondao.NewMemberStorage()
	urequire.False(t, s.IsSealed(), "expect member storage to be not sealed")
	urequire.False(t, r.IsSealed(), "expect voting record to be not sealed")

	_, err := commondao.NewVotingContext(nil, nil)
	urequire.ErrorContains(t, err, "voting record is required")

	_, err = commondao.NewVotingContext(r, nil)
	urequire.ErrorContains(t, err, "member storage is required")

	ctx, err := commondao.NewVotingContext(r, s)
	urequire.NoError(t, err)
	urequire.True(t, ctx.VotingRecord.IsSealed(), "expect voting record to be sealed")
	urequire.True(t, ctx.Members.IsSealed(), "expect member storage to be sealed")
}
