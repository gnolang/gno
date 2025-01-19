package main

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/users"
)

const admin = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

func main() {
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(admin))
	// as admin, grant invites to unregistered user.
	std.TestSetOrigCaller(admin)
	users.GrantInvites(caller.String() + ":1")
	println("done")
}

// Error:
// invalid user g1vdskcmr9wf047h6lta047h6lta047h6lq5m7fe
