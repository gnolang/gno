package main

// SEND: 1000000ugnot

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/archive/boards"
	susers "gno.land/r/sys/users"
)

func main() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	// create a post via registered user
	bid1 := boards.CreateBoard(cross, "test_board1")
	pid := boards.CreateThread(cross, bid1, "First Post (title)", "Body of the first post. (body)")
	bid2 := boards.CreateBoard(cross, "test_board2")

	// create a repost via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(chain.Coins{{"ugnot", 9000000}})

	rid := boards.CreateRepost(cross, bid1, pid, "", "Check this out", bid2)
	println(rid)
	println(boards.Render("test_board1"))
}

// Error:
// please register, otherwise minimum fee 100000000 is required if anonymous
