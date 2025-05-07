package main

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/home"
)

func main() {
	var admin = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	testing.SetOriginCaller(admin)
	cross(home.AdminSetOverride)("Hello World!")
	println("---")
	println(home.Render(""))

	newAdmin := testutils.TestAddress("newAdmin")
	home.Admin.TransferOwnership(newAdmin)
	if err := revive(func() {
		cross(home.AdminSetOverride)("Not admin anymore")
	}); err == nil {
		panic("AdminSetOverride should have aborted the transaction")
	}
}

// Output:
// ---
// Hello World!
