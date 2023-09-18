// PKGPATH: gno.land/r/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"std"
	"strconv"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/groups"
	"gno.land/r/demo/users"
)

var gid groups.GroupID

func main() {
	users.Register("", "gnouser", "my profile")
	gid = groups.CreateGroup("test_group")
	println(gid)

	// delete member via anon user
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test2)
	std.TestSetOrigSend(std.Coins{{"ugnot", 9000000}}, nil)

	groups.DeleteMember(gid, 0)
	println(groups.Render(""))
}

// Output:
// 1
// List of all Groups:
//
//  * [test_group](/r/demo/groups:test_group)
