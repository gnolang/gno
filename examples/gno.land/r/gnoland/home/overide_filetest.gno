package main

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/home"
)

func main() {
	std.TestSetOrigCaller("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
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
// r:  ownable: caller is not owner
