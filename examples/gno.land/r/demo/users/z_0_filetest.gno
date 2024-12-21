package main

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/users"
)

func main() {
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
	std.TestSetOrigSend(std.Coins{std.NewCoin("dontcare", 1)}, nil)
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// incompatible coin denominations: dontcare, ugnot
