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
	users.Register(cross, "main123")

	test1 := testutils.TestAddress("gnouser1")
	testing.SetOriginCaller(test1)
	testing.SetRealm(std.NewUserRealm(test1))
	users.Register(cross, "test123")

	test2 := testutils.TestAddress("gnouser2")
	testing.SetOriginCaller(test2)
	testing.SetRealm(std.NewUserRealm(test2))
	users.Register(cross, "test223")

	test3 := testutils.TestAddress("gnouser3")
	testing.SetOriginCaller(test3)
	testing.SetRealm(std.NewUserRealm(test3))
	users.Register(cross, "test323")

	testing.SetOriginCaller(caller)
	testing.SetRealm(std.NewUserRealm(caller))

	gid = groups.CreateGroup(cross, "test_group")
	println(groups.Render("test_group"))

	groups.AddMember(cross, gid, test2.String(), 42, "metadata3")

	groups.DeleteMember(cross, gid, 0)
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
