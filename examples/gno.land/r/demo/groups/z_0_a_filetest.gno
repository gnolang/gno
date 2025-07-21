// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"gno.land/r/demo/groups"
)

var gid groups.GroupID

func main() {
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	println(groups.Render(""))
}

// Error:
// user not found
