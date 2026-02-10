package main

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	users "gno.land/r/gnoland/users/v1"
)

func main() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")
	bid1 := boards.CreateBoard(cross, "test_board1")
	pid := boards.CreateThread(cross, bid1, "First Post (title)", "Body of the first post. (body)")
	boards.CreateBoard(cross, "test_board2")

	// create a repost to a non-existing destination board
	rid := boards.CreateRepost(cross, bid1, pid, "", "Check this out", 5)
	println(rid)
	println(boards.Render("test_board1"))
}

// Error:
// dst board not exist
