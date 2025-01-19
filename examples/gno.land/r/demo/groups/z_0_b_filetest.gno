// PKGPATH: gno.land/r/demo/groups_test
package groups_test

import (
	"gno.land/r/demo/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	users.Register("gnouser123")
	gid = groups.CreateGroup("test_group")
	println(gid)
	println(groups.Render(""))
}

// Error:
// user not found
