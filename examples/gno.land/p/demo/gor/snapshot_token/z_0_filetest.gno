package main

import (
	"gno.land/p/demo/gor/snapshot_token"
	"gno.land/p/demo/testutils"
	"std"
)

const zeroAddress = std.Address("")

func main() {
	test1 := testutils.TestAddress("test1")
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test1)
	println(test1)

	// mint coin
	amount := uint64(100)

	sp_token := snapshot_token.NewSnapShotToken()

	sp_token.TransferVotingUnits(zeroAddress, test1, amount)

	println("after transferring voting unit:", sp_token.GetAmount(test1))

	lastHeight := std.GetHeight()
	std.TestSkipHeights(1) //skip current block, into voting period

	lastSupply := sp_token.GetPastTotalSupply(lastHeight)
	last2Supply := sp_token.GetPastTotalSupply(lastHeight - 1)
	println("last height total supply: ", lastSupply)
	println("(last height -1) total supply: ", last2Supply)

	lastAmount := sp_token.GetPastAmount(test1, lastHeight)
	last2Amount := sp_token.GetPastAmount(test1, lastHeight-1)
	println("last amount  of test1 is: ", lastAmount)
	println("last 2 amount  of test1 is: ", last2Amount)

	println(sp_token.RenderHome(test1))

	// Tranfer voting units
	sp_token.TransferVotingUnits(test1, test2, amount)
	println(sp_token.RenderHome(test1))
	println("---------------------------------")
	println(sp_token.RenderHome(test2))
}

// Output:
