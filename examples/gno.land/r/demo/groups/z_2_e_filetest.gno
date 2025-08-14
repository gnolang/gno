// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	groups.DeleteGroup(cross, gid)
	println(groups.Render(""))
}

// Error:
// r/sys/users: current realm/user does not exist in whitelist: CodeRealm{ g1njxh4leja7h52ea0lnq9crx3j6782g77nc7yd4, gno.land/r/sys/users } (previous: CodeRealm{ g1zypp3wv4qg9cv9rs57r34jj6z7uuqxjztyeqyk, gno.land/r/gnoland/users/v1 })
