// PKGPATH: gno.land/p/archives/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/archives/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	caller := std.OriginCaller()
	testing.SetRealm(std.NewUserRealm(caller))
	users.Register(cross, "gnouser123")

	gid = groups.CreateGroup(cross, "test_group")

	println(gid)
	groups.DeleteGroup(cross, 20)
	println(groups.Render(""))
}

// Error:
// group id (20) does not exists
