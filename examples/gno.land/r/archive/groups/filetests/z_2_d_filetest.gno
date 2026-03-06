// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/archive/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	users.Register(cross, "gnouser123")
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)

	// delete member via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(chain.Coins{{"ugnot", 9000000}})

	groups.DeleteMember(cross, gid, 0)
	println(groups.Render(""))
}

// Error:
// r/gnoland/users: non-user call
