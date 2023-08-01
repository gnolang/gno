package main

import (
	"std"
	"gno.land/p/demo/governance/gvotes"
	grc20 "gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/testutils"
	"errors"
)

func main() {
	const zeroAddress = std.Address("")
	test1 := testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)
	println(test1)

	// mint coin
	amount := uint64(100)
	token := grc20.NewAdminToken("gov", "GOV", 10)
	token.Mint(test1, amount)
	balance, err := token.BalanceOf(test1)
	if err == nil{
		println("balance of test1 is :", balance)
	}
	votes := gvotes.NewVotes()
	votes.TransferVotingUnits(zeroAddress, test1, amount)
	println(votes.RenderHome(test1))
}

// Output:
// g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
// balance of test1 is : 100
// **Total supply**: 100
// **votes of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: 100
// **delegatee of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7