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
	caller := std.OriginCaller() // main
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
	testing.SetOriginSend(std.Coins{{"dontcare", 1}})
	users.Register(caller, "satoshi", "my other profile")
	// as admin, grant invites to gnouser(again) and satoshi.
	testing.SetOriginCaller(admin)
	users.GrantInvites(caller.String() + ":1\n" + test1.String() + ":1")
	println("done")
}

// Output:
// done
