// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/groups"
	susers "gno.land/r/sys/users"
)

var gid groups.GroupID

func main() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	groups.DeleteGroup(cross, gid)
	println(groups.Render(""))
}

// Output:
// 1
// List of all Groups:
