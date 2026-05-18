// PKGPATH: gno.land/r/archive/groups/filetests/z_2_g
package z_2_g

// SEND: 1000000ugnot

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils/v0"
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

	// delete group via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(chain.Coins{{"ugnot", 9000000}})

	groups.DeleteGroup(cross2(cur), gid)
	println(groups.Render(""))
}

// Error:
// unauthorized to delete group
