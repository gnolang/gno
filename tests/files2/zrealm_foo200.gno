package main

import (
	"std"

	"gno.land/r/foo20"
)

var admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
var manfred = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
var unknown = std.Address("g1u0000000000000000000000000000000000000")

func main() {
	// testing read-only public methods

	println(foo20.TotalSupply())
	println(foo20.BalanceOf(admin))
	println(foo20.BalanceOf(manfred))
	println(foo20.BalanceOf(unknown))
	println(foo20.Allowance(admin, manfred))
	println("done")
}

// Output:
// 10100000000
// 10000000000
// 100000000
// 0
// 0
// done
