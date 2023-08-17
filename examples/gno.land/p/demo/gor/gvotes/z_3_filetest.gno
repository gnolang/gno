package main

import (
	"std"
	"gno.land/p/demo/governance/gvotes"
	grc20 "gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/testutils"
)

func main() {
	const zeroAddress = std.Address("")
	test1 := testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)
	println("address of test1 is: ", test1)

	// mint coin
	amount := uint64(100)
	token := grc20.NewAdminToken("gov", "GOV", 10)
	token.Mint(test1, amount)
	balance, err := token.BalanceOf(test1)
	if err == nil{
		println("balance of token is:", balance)
	}

	votes := gvotes.NewVotes()
	votes.TransferVotingUnits(zeroAddress, test1, amount)
	println(votes.RenderHome(test1))
	// burn
	token.Burn(test1, uint64(30))
	balance, err = token.BalanceOf(test1)
	if err == nil{
		println("balance of token after burn is:", balance)
	}
	println("coin burned, voting power not trasferred")
	println(votes.RenderHome(test1))
	// burn votes power
	votes.TransferVotingUnits(test1, zeroAddress, uint64(30))
	println("coin burned, voting power been trasferred")
	println(votes.RenderHome(test1))
}

// Output:
// address of test1 is:  g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
// balance of token is: 100
// **Total supply**: 100
// **votes of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: 100
// **delegatee of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
//
// balance of token after burn is: 70
// coin burned, voting power not trasferred
// **Total supply**: 100
// **votes of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: 100
// **delegatee of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
//
// coin burned, voting power been trasferred
// **Total supply**: 70
// **votes of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: 70
// **delegatee of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7