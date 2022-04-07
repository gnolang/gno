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
	users.GrantInvites(caller.String() + ":1")
	// switch back to caller
	std.TestSetOrigCaller(caller)
	// invite another addr
	test1 := testutils.TestAddress("test1")
	users.Invite(test1.String())
	// switch to test1
	std.TestSetOrigCaller(test1)
	std.TestSetOrigSend(std.Coins{{"dontcare", 1}}, nil)
	users.Register(caller, "satoshi", "my other profile")
	println(users.Render(""))
}

// Output:
// ## user gnouser (g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4),invites:0,inviter:
//
// my profile
//
// ## user satoshi (g1w3jhxap3ta047h6lta047h6lta047h6l4mfnm7),invites:0,inviter:g17rgsdnfxzza0sdfsdma37sdwxagsz378833ca4
//
// my other profile
