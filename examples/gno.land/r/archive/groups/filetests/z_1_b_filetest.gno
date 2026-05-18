// PKGPATH: gno.land/r/archive/groups/filetests/z_1_b
package z_1_b

// SEND: 1000000ugnot

import (
	"chain/runtime"
	"testing"

	"gno.land/r/archive/groups"
	uinit "gno.land/r/sys/users/init"
)

var gid groups.GroupID

func main(cur realm) {
	caller := runtime.OriginCaller()
	testing.SetRealm(testing.NewUserRealm(caller))
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "gnouser123", caller)
	testing.SetHeight(123)

	gid = groups.CreateGroup(cross2(cur), "test_group")
	println(gid)
	groups.AddMember(cross2(cur), 2, "g1vahx7atnv4erxh6lta047h6lta047h6ll85gpy", 55, "metadata3")
	println(groups.Render(""))
}

// Error:
// group id (2) does not exists
