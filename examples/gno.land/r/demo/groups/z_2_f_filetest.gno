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
	groups.DeleteGroup(20)
	println(groups.Render(""))
}

// Error:
// group id (20) does not exists
