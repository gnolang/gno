// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var bid boards.BoardID

func init() {
	users.Register("", "gnouser", "my profile")

	bid = boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(bid, pid, pid, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Output:
// \[[post](/r/demo/boards?help&__func=CreateThread&bid=1&body.type=textarea)]
//
// ----------------------------------------
// ## [First Post (title)](/r/demo/boards:test_board/1)
//
// Body of the first post. (body)
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm UTC](/r/demo/boards:test_board/1) \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)] (0 replies) (0 reposts)
//
// ----------------------------------------
// ## [Second Post (title)](/r/demo/boards:test_board/2)
//
// Body of the second post. (body)
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm UTC](/r/demo/boards:test_board/2) \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=2&postid=2)] (1 replies) (0 reposts)
