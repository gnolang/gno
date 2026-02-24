// MAXALLOC: 50000000
// SEND: 1000000ugnot

package main

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/archive/boards"
	susers "gno.land/r/sys/users"
)

func main() {
	aliceAddr := testutils.TestAddress("alice")

	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "alice_123", aliceAddr)
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	bid := boards.CreateBoard(cross, "testboard")
	boards.CreateThread(cross, bid, "testthread", "testbody")
	n := 5000
	for i := 0; i < n; i++ {
		boards.CreateReply(cross, bid, boards.PostID(1), boards.PostID(1), "reply")
	}
	println("ok")
}

// Output:
// ok
