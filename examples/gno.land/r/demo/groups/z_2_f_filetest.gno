// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"std"

	"gno.land/r/demo/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	caller := std.OriginCaller()
	std.TestSetRealm(std.NewUserRealm(caller))
	users.Register("gnouser123")

	gid = groups.CreateGroup("test_group")

	println(gid)
	groups.DeleteGroup(20)
	println(groups.Render(""))
}

// Error:
// group id (20) does not exists
