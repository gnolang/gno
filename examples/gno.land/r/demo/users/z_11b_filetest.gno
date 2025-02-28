package main

// SEND: 200000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/users"
)

const admin = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

func main() {
	testing.SetOriginCaller(admin)
	caller := std.OriginCaller() // main
	// add restricted name
	users.AdminAddRestrictedName("superrestricted")
	// grant invite to caller
	users.Invite(caller.String())
	// set back caller
	testing.SetOriginCaller(caller)
	// register restricted name with admin invite
	users.Register(admin, "superrestricted", "my profile")
	println("done")
}

// Output:
// done
