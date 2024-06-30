// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/groups"
	"gno.land/r/demo/users"
)

var gid groups.GroupID

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	caller := std.GetOrigCaller() // main
	users.Register("", "gnouser0", "my profile 1")

	std.TestSetOrigCaller(admin)
	users.GrantInvites(caller.String() + ":1")
	// switch back to caller
	std.TestSetOrigCaller(caller)
	// invite another addr
	test1 := testutils.TestAddress("gnouser1")
	users.Invite(test1.String())
	// switch to test1
	std.TestSetOrigCaller(test1)
	users.Register(caller, "gnouser1", "my other profile 1")

	std.TestSetOrigCaller(admin)
	users.GrantInvites(caller.String() + ":1")
	// switch back to caller
	std.TestSetOrigCaller(caller)
	// invite another addr
	test2 := testutils.TestAddress("gnouser2")
	users.Invite(test2.String())
	// switch to test1
	std.TestSetOrigCaller(test2)
	users.Register(caller, "gnouser2", "my other profile 2")

	std.TestSetOrigCaller(admin)
	users.GrantInvites(caller.String() + ":1")
	// switch back to caller
	std.TestSetOrigCaller(caller)
	// invite another addr
	test3 := testutils.TestAddress("gnouser3")
	users.Invite(test3.String())
	// switch to test1
	std.TestSetOrigCaller(test3)
	users.Register(caller, "gnouser3", "my other profile 3")

	std.TestSetOrigCaller(caller)

	gid = groups.CreateGroup("test_group")
	println(gid)

	groups.AddMember(gid, test3.String(), 32, "i am from UAE")
	println(groups.Render("test_group"))
}

// Output:
// 1
// Group ID: 0000000001
//
// Group Name: test_group
//
// Group Creator: gnouser0
//
// Group createdAt: 2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001
//
// Group Last MemberID: 0000000001
//
// Group Members:
//
// 			[0000000000, g1vahx7atnv4erxh6lta047h6lta047h6ll85gpy, 32, i am from UAE, 2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001],
