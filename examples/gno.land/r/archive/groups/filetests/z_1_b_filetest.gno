// PKGPATH: gno.land/p/archive/groups_test
package groups_test

// SEND: 1000000ugnot

import (
	"chain/runtime"
	"testing"

	"gno.land/r/archive/groups"
	uinit "gno.land/r/sys/users/init"
)

var gid groups.GroupID

func main() {
	caller := runtime.OriginCaller()
	testing.SetRealm(testing.NewUserRealm(caller))
	uinit.RegisterUser(cross, "gnouser123", caller)

	gid = groups.CreateGroup(cross, "test_group")
	println(gid)
	groups.AddMember(cross, 2, "g1vahx7atnv4erxh6lta047h6lta047h6ll85gpy", 55, "metadata3")
	println(groups.Render(""))
}

// Error:
// group id (2) does not exists
