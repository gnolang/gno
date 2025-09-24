// MAXALLOC: 20000
// SEND: 1000000ugnot

package main

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/archive/boards"
	"gno.land/r/gnoland/users/v1"
)

func main() {
	aliceAddr := testutils.TestAddress("alice")

	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(runtime.NewUserRealm(aliceAddr))

	users.Register(cross, "alice_123")

	bid := boards.CreateBoard(cross, "testboard")
	boards.CreateThread(cross, bid, "testthread", "testbody")
	n := 5000
	for i := 0; i < n; i++ {
		boards.CreateReply(cross, bid, boards.PostID(1), boards.PostID(1), "reply")
	}
	println("ok")
}

// allocation limit while importing.

// Error:
// main/alloc_10c_long.gno:8:2-11: allocation limit exceeded
