package main

// SEND: 1000000ugnot

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	bid1 boards.BoardID
	bid2 boards.BoardID
	pid  boards.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(chain.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")

	bid1 = boards.CreateBoard(cross, "test_board1")
	pid = boards.CreateThread(cross, bid1, "First Post (title)", "Body of the first post. (body)")
	bid2 = boards.CreateBoard(cross, "test_board2")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(chain.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	rid := boards.CreateRepost(cross, bid1, pid, "", "Check this out", bid2)
	println(rid)
	println(boards.Render("test_board2"))
}

// Output:
// 1
// \[[post](/r/demo/boards$help&func=CreateThread&bid=2)]
//
// ----------------------------------------
// Repost: Check this out
// ## [First Post (title)](/r/demo/boards:test_board1/1)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm UTC](/r/demo/boards:test_board1/1) \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)] (0 replies) (1 reposts)
//
//
