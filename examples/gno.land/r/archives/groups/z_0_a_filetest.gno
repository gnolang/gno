// PKGPATH: gno.land/r/archives/groups_test
package groups

// SEND: 1000000ugnot

import (
	"gno.land/r/archives/groups"
)

var gid groups.GroupID

func main() {
	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	println(groups.Render(""))
}

// Error:
// user not found
