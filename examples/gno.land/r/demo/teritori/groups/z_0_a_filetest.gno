// PKGPATH: gno.land/r/groups_test
package groups_test

import (
	"gno.land/r/demo/teritori/groups"
)

var gid groups.GroupID

func main() {
	gid = groups.CreateGroup("test_group")
	println(gid)
	println(groups.Render(""))
}

// Output:
// 1
// List of all Groups:
//
//  * [test_group](/r/demo/groups:test_group)
