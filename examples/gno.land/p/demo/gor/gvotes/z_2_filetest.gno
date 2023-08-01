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
		println("balance of token is:", balance)
	}
	votes := gvotes.NewVotes()
	// get vote will panic if not transfering voting units expicitly
	println(votes.RenderHome(test1))
}

// Error:
// Votes: g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7 not exsit!