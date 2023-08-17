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
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test1)
	println("address of test1 is: ", test1)

	// mint coin
	amount := uint64(100)
	token := grc20.NewAdminToken("gov", "GOV", 10)
	token.Mint(test1, amount)
	balance, err := token.BalanceOf(test1)
	if err == nil{
		println("balance of test1 is:", balance)
	}
	votes := gvotes.NewVotes()
	rr := votes.TransferVotingUnits(zeroAddress, test1, amount)

	println("transfer to test2-------------------")
	token.Transfer(test1, test2, amount)
	balance, err = token.BalanceOf(test1)
	if err == nil{
		println("balance of test1 is:", balance)
	}

	err1 := votes.TransferVotingUnits(test1, test2, amount)
	if err1 != nil{
		println(err1.Error())
	}

	println("transfer again---------------------")
	err = token.Transfer(test1, test2, amount)
	if err != nil{
		println(err.Error())
	}
	err = votes.TransferVotingUnits(test1, test2, amount)
	if err != nil{
		println(err.Error())
	}

	// transfer to self
	err = votes.TransferVotingUnits(test1, test1, amount)
	if err != nil{
		println(err.Error())
	}

	println(votes.RenderHome(test1))
	println(votes.RenderHome(test2))
}

// Output:
// address of test1 is:  g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
// balance of test1 is: 100
// transfer to test2-------------------
// balance of test1 is: 0
// transfer again---------------------
// insufficient balance
// insufficient votes
// cannot send transfer to self
// **Total supply**: 100
// **votes of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: 0
// **delegatee of g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 is**: g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7
//
// **Total supply**: 100
// **votes of g1w3jhxapjta047h6lta047h6lta047h6laqcyu4 is**: 100
// **delegatee of g1w3jhxapjta047h6lta047h6lta047h6laqcyu4 is**: g1w3jhxapjta047h6lta047h6lta047h6laqcyu4