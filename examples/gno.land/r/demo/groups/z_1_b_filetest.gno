// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 200000000ugnot

import (
	"gno.land/r/demo/groups"
	"gno.land/r/demo/users"
)

var gid groups.GroupID

func main() {
	users.Register("", "gnouser", "my profile")
	gid = groups.CreateGroup("test_group")
	println(gid)
	groups.AddMember(2, "g1vahx7atnv4erxh6lta047h6lta047h6ll85gpy", 55, "metadata3")
	println(groups.Render(""))
}

// Error:
// group id (2) does not exists
