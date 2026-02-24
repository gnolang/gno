package main

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	susers "gno.land/r/sys/users"
)

var (
	bid1 boards.BoardID
	bid2 boards.BoardID
	pid  boards.PostID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))

	bid1 = boards.CreateBoard(cross, "test_board1")
	pid = boards.CreateThread(cross, bid1, "First Post (title)", "Body of the first post. (body)")
	bid2 = boards.CreateBoard(cross, "test_board2")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	rid := boards.CreateRepost(cross, bid1, pid, "", "Check this out", bid2)
	println(rid)
	println(boards.Render("test_board2"))
}

// Output:
// 1
// \[[post](/r/archive/boards$help&func=CreateThread&bid=2)]
//
// ----------------------------------------
// Repost: Check this out
// ## [First Post (title)](/r/archive/boards:test_board1/1)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm UTC](/r/archive/boards:test_board1/1) \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)] (0 replies) (1 reposts)
//
//
