// PKGPATH: gno.land/r/archives/groups_test
package groups

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/archives/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	println(groups.Render(""))
}

// Output:
// 1
// List of all Groups:
//
//  * [test_group](/r/demo/groups:test_group)
//
