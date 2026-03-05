package main

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/gnoland/home"
)

func main() {
	var admin = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	testing.SetOriginCaller(admin)
	home.AdminSetOverride(cross, "Hello World!")
	println("---")
	println(home.Render(""))

	newAdmin := testutils.TestAddress("newAdmin")
	home.AdminTransferOwnership(cross, newAdmin)
	if err := revive(func() {
		home.AdminSetOverride(cross, "Not admin anymore")
	}); err == nil {
		panic("AdminSetOverride should have aborted the transaction")
	}
}

// Output:
// ---
// Hello World!
