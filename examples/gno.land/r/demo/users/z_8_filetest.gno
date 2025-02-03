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
	testing.SetOriginCaller(std.NewUserRealm(admin))
	users.GrantInvites(caller.String() + ":1")
	// switch back to caller
	testing.SetOriginCaller(std.NewUserRealm(caller))
	// invite another addr
	test1 := testutils.TestAddress("test1")
	users.Invite(test1.String())
	// switch to test1
	testing.SetOriginCaller(std.NewUserRealm(test1))
	testing.SetOriginSend(std.Coins{{"dontcare", 1}}, nil)
	users.Register(caller, "satoshi", "my other profile")
	// as admin, grant invites to gnouser(again) and nonexistent user.
	testing.SetOriginCaller(std.NewUserRealm(admin))
	test2 := testutils.TestAddress("test2")
	users.GrantInvites(caller.String() + ":1\n" + test2.String() + ":1")
	println("done")
}

// Error:
// invalid user g1w3jhxapjta047h6lta047h6lta047h6laqcyu4
