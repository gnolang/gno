package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

func main() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	cross(users.Register)("gnouser123")
	bid1 := cross(boards.CreateBoard)("test_board1")
	pid := cross(boards.CreateThread)(bid1, "First Post (title)", "Body of the first post. (body)")
	cross(boards.CreateBoard)("test_board2")

	// create a repost to a non-existing destination board
	rid := cross(boards.CreateRepost)(bid1, pid, "", "Check this out", 5)
	println(rid)
	println(boards.Render("test_board1"))
}

// Error:
// dst board not exist
