// MAXALLOC: 20000
// SEND: 1000000ugnot

package main

import (
	"std"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/archive/boards"
	"gno.land/r/gnoland/users/v1"
)

func main() {
	aliceAddr := testutils.TestAddress("alice")

	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(std.NewUserRealm(aliceAddr))

	users.Register(cross, "alice_123")

	bid := boards.CreateBoard(cross, "testboard")
	boards.CreateThread(cross, bid, "testthread", "testbody")
	n := 5000
	for i := 0; i < n; i++ {
		boards.CreateReply(cross, bid, boards.PostID(1), boards.PostID(1), "reply")
	}
	println("ok")
}

// allocation limit exceeded during importing.

// Error:
// main/alloc_10c.gno:7:2-7: allocation limit exceeded
