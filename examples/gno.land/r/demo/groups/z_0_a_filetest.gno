// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"gno.land/r/demo/groups"
)

var gid groups.GroupID

func main() {
	gid = groups.CreateGroup("test_group")
	println(gid)
	println(groups.Render(""))
}

// Error:
// missing crossing() after cross call to main from gno.land/r/demo/groups_test to gno.land/r/demo/groups_test
