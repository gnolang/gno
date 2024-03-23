package gnousd

// SEND: 100gnot

import (
	"std"

	"gno.land/p/testutils"
	"gno.land/r/gnousd"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
const test1 = testutils.TestAddress("test1")
const test2 = testutils.TestAddress("test2")
const test3 = testutils.TestAddress("test3")
const test4 = testutils.TestAddress("test4")

func main() {
	caller := std.GetOrigCaller() // main

	println(gnousd.Render(""))

	std.TestSetOrigCaller(admin)
	gnousd.UpdateCoinPrice(1000000)
	gnousd.UpdateSharePrice(100000000)

	println(gnousd.Render(""))

	/*
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
	*/
}

// Output:
// # GnoUSD ($GUSD)
//
// Coin price: 1000000
// Share price: 100000000
// Cycles: 1
// Height (start, current, max): 123, 123, 2123
//
// # GnoUSD ($GUSD)
//
// Coin price: 1000000
// Share price: 100000000
// Cycles: 1
// Height (start, current, max): 123, 123, 2123
