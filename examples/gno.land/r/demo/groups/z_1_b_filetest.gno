// PKGPATH: gno.land/r/demo/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/groups"
	users "gno.land/r/gnoland/users/v1"
)

var gid groups.GroupID

func main() {
	caller := std.OriginCaller()
	testing.SetRealm(std.NewUserRealm(caller))
	users.Register(cross, "gnouser123")

	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	groups.AddMember(cross, 2, "g1vahx7atnv4erxh6lta047h6lta047h6ll85gpy", 55, "metadata3")
	println(groups.Render(""))
}

// Error:
// should not happen, method and object not in same realm
