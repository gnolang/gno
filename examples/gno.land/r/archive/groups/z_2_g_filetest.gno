// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/archive/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	caller := std.OriginCaller()
	testing.SetRealm(std.NewUserRealm(caller))
	users.Register(cross, "gnouser123")

	gid = groups.CreateGroup(cross, "test_group")
	println(gid)

	// delete group via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(std.Coins{{"ugnot", 9000000}})

	groups.DeleteGroup(cross, gid)
	println(groups.Render(""))
}

// Error:
// unauthorized to delete group
