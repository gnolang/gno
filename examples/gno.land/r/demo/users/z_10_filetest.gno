// PKGPATH: gno.land/r/demo/users_test
package users_test

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/users"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func init() {
	caller := std.GetOrigCaller() // main
	test2 := testutils.TestAddress("test2")
	// as admin, invite gnouser and test2
	std.TestSetOrigCaller(admin)
	users.Invite(caller.String() + "\n" + test2.String())
	// register as caller
	std.TestSetOrigCaller(caller)
	users.Register(admin, "gnouser", "my profile")
}

func main() {
	// register as test2
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test2)
	users.Register(admin, "test222", "my profile 2")
	println("done")
}

// Output:
// done
