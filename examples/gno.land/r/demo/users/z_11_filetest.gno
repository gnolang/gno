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
	users.AdminAddRestrictedName("superrestricted")

	// test restricted name
	std.TestSetOrigCaller(caller)
	users.Register("", "superrestricted", "my profile")
	println("done")
}

// Error:
// restricted name: superrestricted
