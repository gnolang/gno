package main

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/home"
)

func main() {
	std.TestSetOrigCaller("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	home.AdminSetOverride("Hello World!")
	println(home.Render(""))
	home.AdminTransferOwnership(testutils.TestAddress("newAdmin"))
	defer func() {
		r := recover()
		println("r: ", r)
	}()
	home.AdminSetOverride("Not admin anymore")
}

// Output:
// Hello World!
// r:  unauthorized; caller is not owner
