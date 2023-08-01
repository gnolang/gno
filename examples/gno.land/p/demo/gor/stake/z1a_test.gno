package stake

import (
	"gno.land/p/demo/avl"
	governor "gno.land/p/demo/governance/governor"
	"gno.land/p/demo/testutils"
	"std"
	"testing"
)

// delegate, vote, success
func TestDelegateVoteSucceed2(t *testing.T) {
	println("----------TestDelegateVoteSucceed2----------")
	env := getEnv()
	d0 := env.delegateAddrs[0]
	d1 := env.delegateAddrs[1]
	d2 := env.delegateAddrs[2]
	delegator := env.delegator
	poolAddr := env.pool
	s := env.stakes[1]

	snapshot := std.GetHeight()
	println("delegator: ", delegator)
	println("snapshot: ", snapshot)
	println("poolAddr: ", poolAddr)

	// this is needed cuz "main" is set to be origPkgAddr while testing, not the real pkgpath in msgCall
	// std.TestSetOrigPkgAddr(std.Address(poolAddr)) // origPkgAddr is set to `main` defaulted
	// test Delegate(), to delegate native coin

	std.TestSetOrigCaller(std.Address(delegator))  

	err := s.Delegate(d0, "FOO", 8, false)
	if err != nil{
		println(err.Error())
	}
	err = s.Delegate(d1, "FOO", 27, false)
	if err != nil{
		println(err.Error())
	}
	err = s.Delegate(d2, "FOO", 71, false)
	if err != nil{
		println(err.Error())
	}

	std.TestSkipHeights(1) //skip current block to
	snapshot = std.GetHeight()

	// check pool balance
	amount := s.GovToken.Balance(std.Address(poolAddr))
	println("pool amount: ", amount)
	assert(t, amount == 106, "pool balance mismatch")

	// get delegation and check
    indexedDelegation, err := s.GetDelegation(delegator, d0)
	assert(t, err == nil, "get indexedDelegation error")
	delegation, err := indexedDelegation.GetDelegationByIndex(d0)
	shares := delegation.GetSnapshotDelegationShares(snapshot)
	assert(t, shares == 8, "delegate shares not right")

	indexedDelegation, err = s.GetDelegation(delegator, d1)
	assert(t, err == nil, "get indexedDelegation error")
	delegation, err = indexedDelegation.GetDelegationByIndex(d1)
	shares = delegation.GetSnapshotDelegationShares(snapshot)
	assert(t, shares == 27, "delegate shares not right")

	indexedDelegation, err = s.GetDelegation(delegator, d2)
	assert(t, err == nil, "get indexedDelegation error")
	delegation, err = indexedDelegation.GetDelegationByIndex(d2)
	shares = delegation.GetSnapshotDelegationShares(snapshot)
	assert(t, shares == 71, "delegate shares not right")

	// ----------------------------------------------------------------
	// gov
	minDeposit := std.Coin{
		Denom:  "ugnot",
		Amount: 64000000,
	}
	gov := governor.NewGovernor("test_gov", s, s.GovToken, 0, 5, minDeposit, 0, 40, 1, 2)

	// Propose
	textContent := "this is a proposal from max"
	callMsg := &governor.CallMsg{
		Caller:  std.GetOrigCaller().String(),
		PkgPath: "gno.land/r/demo/greet",
		Fn:      "hello",
		Args:    []string{"max"},
	}
	proposal := &governor.Proposal{
		TextContent:  textContent,
		CallMsg:      callMsg,
		ProposalKind: governor.ExecutableProposal,
	}
	proposalId := gov.Propose(proposal)
	println("proposed")
	// proposal created, not start voting
	// println("proposal RenderHome:", gov.RenderHome(proposalId))

	// vote
	std.TestSkipHeights(1) //skip current block to voting period

	// For: 8+1=9, 1 is the votingpower the test delegate own
	std.TestSetOrigCaller(std.Address(d0))
	err = gov.CastVoteWithMemo(proposalId, governor.Against, "")
	assert(t, err == nil, "cast vote error")
	// Against: 27+1
	std.TestSetOrigCaller(std.Address(d1))
	err = gov.CastVoteWithMemo(proposalId, governor.Against, "")
	assert(t, err == nil, "cast vote error")

	// Against: 71+1
	std.TestSetOrigCaller(std.Address(d2))
	err = gov.CastVoteWithMemo(proposalId, governor.For, "")
	assert(t, err == nil, "cast vote error")

	// vote end
	// println("proposal RenderHome:", gov.RenderHome(proposalId))

	// set caller
	std.TestSetOrigCaller(std.Address(delegator))

	// skipping voting period--------------------------------
	std.TestSkipHeights(5)

	state, isBurnDeposit, tallyResult, err := gov.Tally(proposalId)  
	if err != nil{
		panic(err.Error())
	}
	assert(t, state == governor.Defeated, "should be succeeded")
	assert(t, tallyResult.GetVotes(governor.For) == 72, "should be 72")
	assert(t, tallyResult.GetVotes(governor.Against) == 37, "should be 37")
	assert(t, tallyResult.GetVotes(governor.Abstain) == 0, "should be 0")
	assert(t, tallyResult.GetVotes(governor.NoWithVeto) == 0, "should be 0")

	println("proposal RenderHome:", gov.RenderHome(proposalId))
}