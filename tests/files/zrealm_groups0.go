// PKGPATH: gno.land/r/groups_test
package groups_test

import (
	"gno.land/p/groups"
)

var group *groups.Group

func init() {
	group = groups.NewGroup("Test Group")
	group.AddPost("First Post", "Test Body")
}

func main() {
	println(group.String())
}

// Output:
// # [group] Test Group
//
// ## First Post
// Test Body
