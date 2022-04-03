package main

// SEND: 2000gnot

import (
	"gno.land/p/testutils"
	"gno.land/r/users"
	"std"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	caller := std.GetOrigCaller() // main
	users.Register("", "gnouser", "my profile")
	// as admin, grant invites to gnouser
	std.TestSetOrigCaller(admin)
	users.GrantInvites(caller + ":1")
	// switch back to caller
	std.TestSetOrigCaller(caller)
	// invite another addr
	test1 := testutils.TestAddress("test1")
	users.Invite(test1)
	// switch to test1
	std.TestSetOrigCaller(test1)
	std.TestSetTxSend(std.Coins{{"dontcare", 1}})
	users.Register(caller, "satoshi", "my other profile")
	// as admin, grant invites to gnouser(again) and satoshi.
	std.TestSetOrigCaller(admin)
	users.GrantInvites(caller + ":1\n" + test1 + ":1")
	println("done")
}

// Output:
// done
