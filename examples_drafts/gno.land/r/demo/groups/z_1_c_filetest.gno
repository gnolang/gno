// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/groups"
)

var gid groups.GroupID

func main() {
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)

	// add member via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(std.Coins{{"ugnot", 9000000}})

	groups.AddMember(cross, gid, test2.String(), 42, "metadata3")
}

// Error:
// user not found
