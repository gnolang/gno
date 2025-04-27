package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

func init() {
	// register
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	cross(users.Register)("gnouser123")

	// create board and post
	bid := cross(boards.CreateBoard)("test_board")
	cross(boards.CreateThread)(bid, "First Post (title)", "Body of the first post. (body)")
}

func main() {
	println(boards.Render("test_board"))
}

// Output:
// \[[post](/r/demo/boards$help&func=CreateThread&bid=1)]
//
// ----------------------------------------
// ## [First Post (title)](/r/demo/boards:test_board/1)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm UTC](/r/demo/boards:test_board/1) \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)] (0 replies) (0 reposts)
//
//
