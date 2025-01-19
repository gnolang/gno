// PKGPATH: gno.land/r/demo/groups_test
package groups_test

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

// Error:
// payment must not be less than 20000000
