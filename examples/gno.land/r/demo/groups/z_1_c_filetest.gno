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
	gid = groups.CreateGroup("test_group")
	println(gid)

	// add member via anon user
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test2)
	std.TestSetOrigSend(std.Coins{{"ugnot", 9000000}}, nil)

	groups.AddMember(gid, test2.String(), 42, "metadata3")
}

// Error:
// unauthorized
