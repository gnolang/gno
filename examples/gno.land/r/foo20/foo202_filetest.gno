package main

// SEND: 100gnot

import (
	"std"

	"gno.land/p/testutils"
	"gno.land/r/foo20"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
const test1 = testutils.TestAddress("test1")
const test2 = testutils.TestAddress("test2")
const test3 = testutils.TestAddress("test3")
const test4 = testutils.TestAddress("test4")

func main() {
	caller := std.GetOrigCaller() // main
	printBalances()

	std.TestSetOrigCaller(admin)
	foo20.Mint(test1, 100000)
	printBalances()

	std.TestSetOrigCaller(test1)
	foo20.Transfer(test2, 1000)
	printBalances()

	std.TestSetOrigCaller(admin)
	foo20.Transfer(test2, 1000)
	printBalances()

	std.TestSetOrigCaller(test2)
	foo20.Transfer(test3, 100)
	printBalances()

	std.TestSetOrigCaller(test1)
	println(foo20.Allowance(test1, test3))
	foo20.Approve(test3, 10000)
	println(foo20.Allowance(test1, test3))
	printBalances()

	std.TestSetOrigCaller(test3)
	foo20.TransferFrom(test1, test4, 10000)
	printBalances()
}

func printBalances() {
	println(
		foo20.BalanceOf(admin),
		foo20.BalanceOf(test1),
		foo20.BalanceOf(test2),
		foo20.BalanceOf(test3),
		foo20.BalanceOf(test4),
	)
}

// Output:
// 10000000000 0 0 0 0
// 10000000000 100000 0 0 0
// 10000000000 99000 1000 0 0
// 9999999000 99000 2000 0 0
// 9999999000 99000 1900 100 0
// 0
// 10000
// 9999999000 99000 1900 100 0
// 9999999000 89000 1900 100 10000
