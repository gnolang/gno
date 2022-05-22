package main

import (
	"std"

	"gno.land/p/testutils"
	"gno.land/r/foo20"
)

const (
	admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	test1 = testutils.TestAddress("test1")
	test2 = testutils.TestAddress("test2")
	test3 = testutils.TestAddress("test3")
	test4 = testutils.TestAddress("test4")
)

func init() {
	std.TestSetOrigCaller(admin)
	foo20.Transfer(test1, 1000)
	foo20.Transfer(test2, 1000)
}

func main() {
	// testing errors.

	printBalances()

	// balance of invalid address
	rec(func() {
		println(foo20.BalanceOf("")) // invalid address
	})

	// transfer of a fresh account
	rec(func() {
		std.TestSetOrigCaller(test3)
		foo20.Transfer(test4, 1) // insufficient balance
	})

	// transfer too much
	rec(func() {
		std.TestSetOrigCaller(test1)
		foo20.Transfer(test4, 1001) // insufficient balance
	})

	// transferFrom without allowance
	rec(func() {
		std.TestSetOrigCaller(test3)
		foo20.TransferFrom(test1, test4, 1) // insufficient allowance
	})

	// transferFrom without enough allowance
	std.TestSetOrigCaller(test1)
	foo20.Approve(test3, 10)
	rec(func() {
		std.TestSetOrigCaller(test3)
		foo20.TransferFrom(test1, test4, 11) // insufficient allowance
	})

	// TODO
	// burn too much
	// mint too much (overflow)

	printBalances()
	println("done")
}

func assertBalanceOfInvalidAddress() {
}

func rec(handler func()) {
	defer func() {
		r := recover()
		if r != nil {
			println(r)
		}
	}()
	handler()
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
// 9999998000 1000 1000 0 0
// invalid address
// insufficient balance
// insufficient balance
// insufficient allowance
// insufficient allowance
// 9999998000 1000 1000 0 0
// done
