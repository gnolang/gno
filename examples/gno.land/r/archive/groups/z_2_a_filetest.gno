// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/archive/groups"
	susers "gno.land/r/sys/users"
)

var gid groups.GroupID

const admin = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")

func main() {
	caller := runtime.OriginCaller() // main123
	testing.SetRealm(testing.NewUserRealm(caller))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "main123", caller)
	testing.SetRealm(testing.NewUserRealm(caller))

	test1 := testutils.TestAddress("gnouser1")
	testing.SetOriginCaller(test1)
	testing.SetRealm(testing.NewUserRealm(test1))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "test123", test1)
	testing.SetRealm(testing.NewUserRealm(test1))

	test2 := testutils.TestAddress("gnouser2")
	testing.SetOriginCaller(test2)
	testing.SetRealm(testing.NewUserRealm(test2))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "test223", test2)
	testing.SetRealm(testing.NewUserRealm(test2))

	test3 := testutils.TestAddress("gnouser3")
	testing.SetOriginCaller(test3)
	testing.SetRealm(testing.NewUserRealm(test3))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "test323", test3)
	testing.SetRealm(testing.NewUserRealm(test3))

	testing.SetOriginCaller(caller)
	testing.SetRealm(testing.NewUserRealm(caller))

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
// Group createdAt: 2009-02-13 23:31:30 +0000 UTC
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
// Group createdAt: 2009-02-13 23:31:30 +0000 UTC
//
// Group Last MemberID: 0000000001
//
// Group Members:
