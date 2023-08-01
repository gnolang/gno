package main

import (
	"gno.land/p/demo/governance/governor"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/foo_governor"
	"std"
)

func main() {
	// test1 -> test2
	test1 := testutils.TestAddress("test1")
	println("test1:", test1)
	test2 := testutils.TestAddress("test2")
	println("test2:", test2)
	test3 := testutils.TestAddress("test3")
	println("test3:", test3)
	println("============================")

	std.TestSetOrigCaller(test1)
	// mint
	println("mint")
	foo_governor.Mint(test1, 10000)
	totalSupply := foo_governor.TotalSupply()
	println("total supply:", totalSupply)

	// tranfer
	foo_governor.Transfer(test1, test2, 3000)

	// // balance of
	balance, err := foo_governor.BalanceOf(test1)
	if err != nil {
		println("balance of test1 is :", balance)
	}
	balance, err = foo_governor.BalanceOf(test2)
	if err != nil {
		println("balance of test2 is :", balance)
	}

	std.TestSkipHeights(1) //skip 1 step, we can only get vote from past block

	// getVote
	height := std.GetHeight()
	println("height is:", height)
	votes_test1 := foo_governor.GetVotes(test1, height-1)
	println(ufmt.Sprintf("test1 votes at block: %d is %d: ", height-1, votes_test1))
	votes_test2 := foo_governor.GetVotes(test2, height-1)
	println(ufmt.Sprintf("test2 votes at block: %d is %d: ", height-1, votes_test2))

	// Propose
	textContent := "this is a proposal from max"
	callMsg := &governor.CallMsg{
		Caller:  std.GetOrigCaller().String(),
		PkgPath: "gno.land/r/demo/greet",
		Fn:      "hello",
		Args:    []string{"max"},
	}
	proposal := &governor.Proposal{
		TextContent: textContent,
		CallMsg: callMsg,
		ProposalKind: governor.ExecutableProposal,
	}
	proposalId := foo_governor.Propose(proposal)
	// println(string(proposalId))

	// println("before voting----------------------------------")
	println("proposal RenderHome:", foo_governor.RenderHome(proposalId))

	// vote
	std.TestSkipHeights(1) //skip current block
	weight := foo_governor.CastVote(proposalId, governor.For)

	println("voted----------------------------------")
	println("weight is: ", weight)
	println("proposal RenderHome:", foo_governor.RenderHome(proposalId))

	// check result
	std.TestSkipHeights(5)
	println("after voting period ----------------------------------")
	println("proposal RenderHome:", foo_governor.RenderHome(proposalId))
}

// Output:
// test1: g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
// test2: g1w3jhxapjta047h6lta047h6lta047h6laqcyu4
// test3: g1w3jhxapnta047h6lta047h6lta047h6lzfhfxt
// ============================
// mint
// total supply: 10000
// height is: 124
// test1 votes at block: 123 is 7000:
// test2 votes at block: 123 is 3000:
// slice[0x7B225456223A7B2254223A7B22506B6750617468223A22676E6F2E6C616E642F702F64656D6F2F676F7665726E616E63652F676F7665726E6F72222C224E616D65223A2250726F706F73616C222C2242617365223A7B22506B6750617468223A22676E6F2E6C616E642F702F64656D6F2F676F7665726E616E63652F676F7665726E6F72222C224669656C6473223A5B7B224E616D65223A2254657874436F6E74656E74222C2254797065223A31362C22456D626564646564223A66616C73652C22546167223A226A736F6E3A5C22746578745F636F6E74656E742C6F6D6974656D7074795C22227D2C7B224E616D65223A2250726F706F73616C4B696E6422...(4386)]
// done propose
// proposal RenderHome: * **Proposal state is***: Pending
// * **Proposal snapShot is***: 124
// * **Proposal deadLine is***: 129
// * **Proposal deposit is***: 0ugnot
//
// voted----------------------------------
// weight is:  7000
// proposal RenderHome: * **Proposal state is***: Active
// * **Proposal snapShot is***: 124
// * **Proposal deadLine is***: 129
// * **Proposal deposit is***: 0ugnot
//
// after voting period ----------------------------------
// proposal RenderHome: * **Proposal state is***: Succeeded
// * **Proposal snapShot is***: 124
// * **Proposal deadLine is***: 129
// * **Proposal deposit is***: 0ugnot