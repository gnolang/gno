// PKGPATH: gno.land/r/test/alloc_10b
// MAXALLOC: 20000000
// SEND: 1000000ugnot

package alloc_10b

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

func main(cur realm) {
	aliceAddr := testutils.TestAddress("alice")

	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(testing.NewUserRealm(aliceAddr))
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "alice_123", aliceAddr)
	testing.SetHeight(123)

	bid := boards.CreateBoard(cross2(cur), "testboard")
	boards.CreateThread(cross2(cur), bid, "testthread", "testbody")
	n := 5000
	for i := 0; i < n; i++ {
		boards.CreateReply(cross2(cur), bid, boards.PostID(1), boards.PostID(1), "reply")
	}
	println("ok")
}

// Error:
// allocation limit exceeded
