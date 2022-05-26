package main

// SEND: 100gnot

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

func main() {
	caller := std.GetOrigCaller() // main

	std.TestSetOrigCaller(admin)
	foo20.Mint(test1, 100000)
	foo20.Mint(test2, 1000000)
	foo20.Mint(test3, 10000000)

	println("----- /")
	println(foo20.Render(""))

	println("----- /balance/<test1>")
	println(foo20.Render("balance/" + test1.String()))

	println("----- /balance/<unknown>")
	println(foo20.Render("balance/g1ublahblahblah"))

	println("----- /404")
	println(foo20.Render("404"))
}

// Output:
// ----- /
// # Foo ($FOO)
//
// * **Decimals**: 4
// * **Total supply**: 10111100000
// * **Known accounts**: 5
//
// ----- /balance/<test1>
// 100000
//
// ----- /balance/<unknown>
// 0
//
// ----- /404
// 404
