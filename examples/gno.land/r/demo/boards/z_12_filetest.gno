// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var (
	bid1 boards.BoardID
	bid2 boards.BoardID
	pid  boards.PostID
)

func init() {
	users.Register("", "gnouser", "my profile")

	bid1 = boards.CreateBoard("test_board1")
	pid = boards.CreateThread(bid1, "First Post (title)", "Body of the first post. (body)")
	bid2 = boards.CreateBoard("test_board2")
}

func main() {
	rid := boards.CreateRepost(bid1, pid, "", "Check this out", bid2)
	println(rid)
	println(boards.Render("test_board2"))
}

// Output:
// 1
// \[[post](/r/demo/boards?help&__func=CreateThread&bid=2&body.type=textarea)]
//
// ----------------------------------------
// Repost: Check this out
// ## [First Post (title)](/r/demo/boards:test_board1/1)
//
// Body of the first post. (body)
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm UTC](/r/demo/boards:test_board1/1) \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)] (0 replies) (1 reposts)
