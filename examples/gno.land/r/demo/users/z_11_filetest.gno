package main

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/r/demo/users"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	std.TestSetOrigCaller(admin)
	users.AdminAddRestrictedName("superrestricted")

	// test restricted name
	users.Register("", "superrestricted", "my profile")
	println("done")
}

// Error:
// restricted name
