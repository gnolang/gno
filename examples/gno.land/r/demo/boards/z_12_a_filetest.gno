package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

func main() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	cross(users.Register)("gnouser123")
	// create a post via registered user
	bid1 := cross(boards.CreateBoard)("test_board1")
	pid := cross(boards.CreateThread)(bid1, "First Post (title)", "Body of the first post. (body)")
	bid2 := cross(boards.CreateBoard)("test_board2")

	// create a repost via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(std.Coins{{"ugnot", 9000000}})

	rid := cross(boards.CreateRepost)(bid1, pid, "", "Check this out", bid2)
	println(rid)
	println(boards.Render("test_board1"))
}

// Error:
// please register, otherwise minimum fee 100000000 is required if anonymous
