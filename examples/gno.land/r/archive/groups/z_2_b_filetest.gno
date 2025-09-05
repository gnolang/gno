// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"gno.land/r/archive/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	users.Register(cross, "gnouser123")
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	groups.DeleteMember(cross, 2, 0)
	println(groups.Render(""))
}

// Error:
// r/gnoland/users: non-user call
