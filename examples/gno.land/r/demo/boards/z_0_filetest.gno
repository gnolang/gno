package main

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var bid boards.BoardID

func init() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")

	bid = boards.CreateBoard(cross, "test_board")
	boards.CreateThread(cross, bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreateThread(cross, bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(cross, bid, pid, pid, "Reply of the second post")
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
// ----------------------------------------
// ## [Second Post (title)](/r/demo/boards:test_board/2)
//
// Body of the second post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm UTC](/r/demo/boards:test_board/2) \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=2&threadid=2)] (1 replies) (0 reposts)
//
//
