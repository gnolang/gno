package main

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var bid boards.BoardID

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	uinit.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))

	bid = boards.CreateBoard(cross, "test_board")
	boards.CreateThread(cross, bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreateThread(cross, bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(cross, bid, pid, pid, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Output:
// \[[post](/r/archive/boards$help&func=CreateThread&bid=1)]
//
// ----------------------------------------
// ## [First Post (title)](/r/archive/boards:test_board/1)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm UTC](/r/archive/boards:test_board/1) \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)] (0 replies) (0 reposts)
//
// ----------------------------------------
// ## [Second Post (title)](/r/archive/boards:test_board/2)
//
// Body of the second post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm UTC](/r/archive/boards:test_board/2) \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=2&threadid=2)] (1 replies) (0 reposts)
//
//
