package params

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/r/gov/dao"
)

// TestProposeSetRunSubmittersExecutesTheListItChecked covers the installed list matching the checked and shown one.
func TestProposeSetRunSubmittersExecutesTheListItChecked(cur realm, t *testing.T) {
	before := GetRunSubmitters()
	if before == nil {
		before = []string{}
	}
	defer testing.SetSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey, before)

	member := g1user
	ally := testutils.TestAddress("ally")
	outlaw := testutils.TestAddress("outlaw")

	testing.SetRealm(testing.NewUserRealm(member))

	addrs := []string{member.String(), ally.String()}
	pr := ProposeSetRunSubmitters(cross(cur), addrs)
	id := dao.MustCreateProposal(cross(cur), pr)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(id)))

	// After the vote, before execution: the caller rewrites its array.
	addrs[1] = outlaw.String()

	dao.ExecuteProposal(cross(cur), id)

	got := GetRunSubmitters()
	uassert.Equal(t, 2, len(got))
	uassert.True(t, contains(got, member.String()))
	uassert.True(t, contains(got, ally.String()))
	uassert.False(t, contains(got, outlaw.String()))
}
