// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"gno.land/r/archive/groups"
)

var gid groups.GroupID

func main(cur realm) {
	gid = groups.CreateGroup(cross2(cur), "test_group")
	println(gid)
	println(groups.Render(""))
}

// Error:
// user not found
