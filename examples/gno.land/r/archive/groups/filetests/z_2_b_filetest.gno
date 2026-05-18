// PKGPATH: gno.land/r/archive/groups/filetests/z_2_b
package z_2_b

// SEND: 200000000ugnot

import (
	"testing"

	"gno.land/r/archive/groups"
	uinit "gno.land/r/sys/users/init"
)

var gid groups.GroupID

func main(cur realm) {
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)
	gid = groups.CreateGroup(cross2(cur), "test_group")
	println(gid)
	groups.DeleteMember(cross2(cur), 2, 0)
	println(groups.Render(""))
}

// Output:
// 1

// Error:
// group id (2) does not exists
