// PKGPATH: gno.land/r/demo/groups_test
package groups_test

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

const admin = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

func main() {
	caller := std.OriginCaller() // main
	std.TestSetRealm(std.NewUserRealm(caller))
	users.Register("main123")

	test1 := testutils.TestAddress("gnouser1")
	std.TestSetOriginCaller(test1)
	std.TestSetRealm(std.NewUserRealm(test1))
	users.Register("test123")

	test2 := testutils.TestAddress("gnouser2")
	std.TestSetOriginCaller(test2)
	std.TestSetRealm(std.NewUserRealm(test2))
	users.Register("test223")

	test3 := testutils.TestAddress("gnouser3")
	std.TestSetOriginCaller(test3)
	std.TestSetRealm(std.NewUserRealm(test3))
	users.Register("test323")

	std.TestSetOriginCaller(caller)
	std.TestSetRealm(std.NewUserRealm(caller))

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
// Group Creator: main123
//
// Group createdAt: 2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001
//
// Group Last MemberID: 0000000001
//
// Group Members:
//
// 			[0000000000, g1vahx7atnv4erxh6lta047h6lta047h6ll85gpy, 32, i am from UAE, 2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001],
//
//
