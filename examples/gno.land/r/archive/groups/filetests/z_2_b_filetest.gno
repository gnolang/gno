// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"gno.land/r/archive/groups"
	uinit "gno.land/r/sys/users/init"
)

var gid groups.GroupID

func main() {
	uinit.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	groups.DeleteMember(cross, 2, 0)
	println(groups.Render(""))
}

// Output:
// 1

// Error:
// group id (2) does not exists
