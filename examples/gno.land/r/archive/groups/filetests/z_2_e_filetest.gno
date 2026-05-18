// PKGPATH: gno.land/r/archive/groups/filetests/z_2_e
package z_2_e

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/groups"
	uinit "gno.land/r/sys/users/init"
)

var gid groups.GroupID

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)
	gid = groups.CreateGroup(cross2(cur), "test_group")
	println(gid)
	groups.DeleteGroup(cross2(cur), gid)
	println(groups.Render(""))
}

// Output:
// 1
// List of all Groups:
