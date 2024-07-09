package main

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/r/demo/users"
)

const admin = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")

func main() {
	caller := std.GetOrigCaller() // main
	std.TestSetOrigCaller(admin)
	// add restricted name
	users.AdminAddRestrictedName("superrestricted")
	// grant invite to caller
	users.Invite(caller.String())
	// set back caller
	std.TestSetOrigCaller(caller)
	// register restricted name with admin invite
	users.Register(admin, "superrestricted", "my profile")
	println("done")
}

// Output:
// done
