// PKGPATH: gno.land/r/archive/groups/filetests/z_0_a
package z_0_a

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
