// PKGPATH: gno.land/r/test/alloc_10c
// MAXALLOC: 20000
// SEND: 1000000ugnot

package alloc_10c

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/archive/boards"
	users "gno.land/r/sys/users/init"
)

func main(cur realm) {
	aliceAddr := testutils.TestAddress("alice_123")

	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	users.RegisterUser(cross2(cur), "alice_123", aliceAddr)

	bid := boards.CreateBoard(cross2(cur), "testboard")
	boards.CreateThread(cross2(cur), bid, "testthread", "testbody")
	n := 5000
	for i := 0; i < n; i++ {
		boards.CreateReply(cross2(cur), bid, boards.PostID(1), boards.PostID(1), "reply")
	}
	println("ok")
}

// allocation limit exceeded during importing.

// Error:
// gno.land/r/archive/boards/board.gno:4:2-11: allocation limit exceeded (no GC)
