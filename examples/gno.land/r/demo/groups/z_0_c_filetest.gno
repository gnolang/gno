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
	test := testutils.TestAddress("test")
	std.TestSetRealm(std.NewUserRealm(test))
	users.Register("", "gnouser", "my profile")
	gid = groups.CreateGroup("test_group")
	println(gid)
	println(groups.Render(""))
}

// Output:
// 1
// List of all Groups:
//
//  * [test_group](/r/demo/groups:test_group)
//
