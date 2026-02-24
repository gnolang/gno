// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"chain/runtime"
	"testing"

	"gno.land/r/archive/groups"
	susers "gno.land/r/sys/users"
)

var gid groups.GroupID

func main() {
	caller := runtime.OriginCaller()
	testing.SetRealm(testing.NewUserRealm(caller))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "gnouser123", caller)
	testing.SetRealm(testing.NewUserRealm(caller))

	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	groups.AddMember(cross, 2, "g1vahx7atnv4erxh6lta047h6lta047h6ll85gpy", 55, "metadata3")
	println(groups.Render(""))
}

// Error:
// group id (2) does not exists
