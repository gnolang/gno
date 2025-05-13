// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	cross(users.Register)("gnouser123")
	gid = cross(groups.CreateGroup)("test_group")
	println(gid)

	// delete member via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(std.Coins{{"ugnot", 9000000}})

	cross(groups.DeleteMember)(gid, 0)
	println(groups.Render(""))
}

// Error:
// r/gnoland/users: non-user call
