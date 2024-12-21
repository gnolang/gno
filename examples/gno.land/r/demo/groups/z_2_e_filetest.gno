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

func main() {
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
	users.Register("", "gnouser", "my profile")
	gid = groups.CreateGroup("test_group")
	println(gid)
	groups.DeleteGroup(gid)
	println(groups.Render(""))
}

// Output:
// 1
// List of all Groups:
//
//
