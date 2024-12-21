// PKGPATH: gno.land/r/demo/groups_test
package groups_test

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/groups"
)

var gid groups.GroupID

func main() {
	test := testutils.TestAddress("test")
	std.TestSetRealm(std.NewUserRealm(test))
	gid = groups.CreateGroup("test_group")
	println(gid)
	println(groups.Render(""))
}

// Error:
// user not found
