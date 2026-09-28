package params

import (
	"testing"

	prms "sys/params"

	"gno.land/p/nt/uassert/v0"
	"gno.land/r/gov/dao"
)

// TestSysParamSliceFactoriesExecuteTheValueTheyTook covers each slice factory executing the value it was built with, not a rewrite of the caller's array made after the vote.
func TestSysParamSliceFactoriesExecuteTheValueTheyTook(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(g1user))
	// Create, vote, rewrite the caller's array, execute.
	run := func(build func() dao.ProposalRequest, rewrite func()) {
		id := dao.MustCreateProposal(cross(cur), build())
		dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(id)))
		rewrite()
		dao.ExecuteProposal(cross(cur), id)
	}
	strs := func(k string) []string {
		v, _ := prms.GetSysParamStrings("zalias", "p", k)
		return v
	}

	b := []byte("approved")
	run(func() dao.ProposalRequest { return NewSysParamBytesPropRequest(cur, "zalias", "p", "b", b) },
		func() { copy(b, []byte("REWRITE!")) })
	got, _ := prms.GetSysParamBytes("zalias", "p", "b")
	uassert.Equal(t, "approved", string(got))

	s1 := []string{"approved"}
	run(func() dao.ProposalRequest { return NewSysParamStringsPropRequest(cur, "zalias", "p", "s1", s1) },
		func() { s1[0] = "REWRITE" })
	uassert.Equal(t, "approved", strs("s1")[0])

	s2 := []string{"approved"}
	run(func() dao.ProposalRequest {
		return NewSysParamStringsPropRequestWithTitle(cur, "zalias", "p", "s2", "t", s2)
	}, func() { s2[0] = "REWRITE" })
	uassert.Equal(t, "approved", strs("s2")[0])

	s3 := []string{"approved"}
	run(func() dao.ProposalRequest {
		return NewSysParamStringsPropRequestAddWithTitle(cur, "zalias", "p", "s3", "t", s3)
	}, func() { s3[0] = "REWRITE" })
	uassert.Equal(t, "approved", strs("s3")[0])

	// Remove: the approved entry goes, the rewritten one stays.
	testing.SetSysParamStrings("zalias", "p", "s4", []string{"keep", "approved", "REWRITE"})
	s4 := []string{"approved"}
	run(func() dao.ProposalRequest {
		return NewSysParamStringsPropRequestRemoveWithTitle(cur, "zalias", "p", "s4", "t", s4)
	}, func() { s4[0] = "REWRITE" })
	uassert.False(t, contains(strs("s4"), "approved"))
	uassert.True(t, contains(strs("s4"), "REWRITE"))
}
