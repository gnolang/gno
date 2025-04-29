// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

const admin = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

func main() {
	caller := std.OriginCaller() // main123
	testing.SetRealm(std.NewUserRealm(caller))
	cross(users.Register)("main123")

	test1 := testutils.TestAddress("gnouser1")
	testing.SetOriginCaller(test1)
	testing.SetRealm(std.NewUserRealm(test1))
	cross(users.Register)("test123")

	test2 := testutils.TestAddress("gnouser2")
	testing.SetOriginCaller(test2)
	testing.SetRealm(std.NewUserRealm(test2))
	cross(users.Register)("test223")

	test3 := testutils.TestAddress("gnouser3")
	testing.SetOriginCaller(test3)
	testing.SetRealm(std.NewUserRealm(test3))
	cross(users.Register)("test323")

	testing.SetOriginCaller(caller)
	testing.SetRealm(std.NewUserRealm(caller))

	gid = cross(groups.CreateGroup)("test_group")
	println(groups.Render("test_group"))

	cross(groups.AddMember)(gid, test2.String(), 42, "metadata3")

	cross(groups.DeleteMember)(gid, 0)
	println(groups.RenderGroup(gid))
}

// Output:
// Group ID: 0000000001
//
// Group Name: test_group
//
// Group Creator: main123
//
// Group createdAt: 2009-02-13 23:31:30 +0000 UTC m=+1234567890.000000001
//
// Group Last MemberID: 0000000000
//
// Group Members:
//
//
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
//
