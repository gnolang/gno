package main

// SEND: 19900000ugnot

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/users"
)

func main() {
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// payment must not be less than 20000000
