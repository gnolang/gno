package main

// SEND: 200000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/users"
)

const admin = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

func main() {
	caller := std.GetOrigCaller() // main
	testing.SetPreviousRealm(std.NewUserRealm(caller))
	users.AdminAddRestrictedName("superrestricted")

	// test restricted name
	testing.SetPreviousRealm(std.NewUserRealm(caller))
	users.Register("", "superrestricted", "my profile")
	println("done")
}

// Error:
// restricted name: superrestricted
