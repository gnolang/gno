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
	users.Register("gnouser123")
	gid = groups.CreateGroup("test_group")
	println(gid)

	// delete member via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(std.Coins{{"ugnot", 9000000}})

	groups.DeleteMember(gid, 0)
	println(groups.Render(""))
}

// Error:
// missing cross before external crossing() in Register from gno.land/r/demo/groups_test to gno.land/r/gnoland/users/v1
