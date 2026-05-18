// PKGPATH: gno.land/r/archive/groups/filetests/z_2_f
package z_2_f

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
	groups.DeleteGroup(cross2(cur), 20)
	println(groups.Render(""))
}

// Error:
// group id (20) does not exists
