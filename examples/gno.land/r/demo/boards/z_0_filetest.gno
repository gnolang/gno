// PKGPATH: gno.land/r/demo/boards_test
package boards_test

import (
	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
	"std"
)

var bid boards.BoardID

func init() {
	std.TestSetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	err := users.Register("gnouser123")

	bid = boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(bid, pid, pid, "Reply of the second post")
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
// \- [@gnouser123](/r/gnoland/users/v1:gnouser123), [2009-02-13 11:31pm UTC](/r/demo/boards:test_board/1) \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=1&postid=1)] (0 replies) (0 reposts)
//
// ----------------------------------------
// ## [Second Post (title)](/r/demo/boards:test_board/2)
//
// Body of the second post. (body)
// \- [@gnouser123](/r/gnoland/users/v1:gnouser123), [2009-02-13 11:31pm UTC](/r/demo/boards:test_board/2) \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=2&postid=2)] (1 replies) (0 reposts)
//
//
