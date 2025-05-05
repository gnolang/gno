// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"gno.land/r/demo/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	cross(users.Register)("gnouser123")
	gid = cross(groups.CreateGroup)("test_group")
	println(gid)
	cross(groups.DeleteMember)(2, 0)
	println(groups.Render(""))
}

// Error:
// r/gnoland/users: non-user call
