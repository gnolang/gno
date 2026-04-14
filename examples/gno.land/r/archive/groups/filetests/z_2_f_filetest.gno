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
	groups.DeleteGroup(cross, 20)
	println(groups.Render(""))
}

// Error:
// group id (20) does not exists
