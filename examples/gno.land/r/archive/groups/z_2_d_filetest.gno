// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/archive/groups"
	uinit "gno.land/r/sys/users/init"
)

var gid groups.GroupID

func main() {
	uinit.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)

	// delete member via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(chain.Coins{{"ugnot", 9000000}})

	groups.DeleteMember(cross, gid, 0)
	println(groups.Render(""))
}

// Output:
// 1

// Error:
// unauthorized to delete member
