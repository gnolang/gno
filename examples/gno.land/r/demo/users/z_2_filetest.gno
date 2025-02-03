package main

// SEND: 200000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/users"
)

const admin = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

func main() {
	caller := std.GetOrigCaller() // main
	users.Register("", "gnouser", "my profile")
	// as admin, grant invites to gnouser
	testing.SetOriginCaller(admin)
	users.GrantInvites(caller.String() + ":1")
	// switch back to caller
	testing.SetOriginCaller(caller)
	// invite another addr
	test1 := testutils.TestAddress("test1")
	users.Invite(test1.String())
	// switch to test1
	testing.SetOriginCaller(test1)
	users.Register(caller, "satoshi", "my other profile")
	println("done")
}

// Output:
// done
