package main

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/gnoland/home"
)

func main() {
	var admin = address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh")
	testing.SetOriginCaller(admin)
	home.AdminSetOverride(cross, "Hello World!")
	println("---")
	println(home.Render(""))

	newAdmin := testutils.TestAddress("newAdmin")
	home.Admin.TransferOwnership(newAdmin)
	if err := revive(func() {
		home.AdminSetOverride(cross, "Not admin anymore")
	}); err == nil {
		panic("AdminSetOverride should have aborted the transaction")
	}
}

// Output:
// ---
// Hello World!
