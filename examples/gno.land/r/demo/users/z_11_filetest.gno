package main

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/r/demo/users"
)

const admin = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

func main() {
	caller := std.GetOrigCaller() // main
	std.TestSetOrigCaller(admin)
	users.AdminAddRestrictedName("superrestricted")

	// test restricted name
	std.TestSetOrigCaller(caller)
	users.Register("", "superrestricted", "my profile")
	println("done")
}

// Error:
// restricted name: superrestricted
