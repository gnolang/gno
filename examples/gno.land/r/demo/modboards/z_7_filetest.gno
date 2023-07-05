// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 2000000000ugnot

import (
	"gno.land/r/demo/modboards"
	"gno.land/r/demo/users"
)

func init() {
	// register
	users.Register("", "gnouser", "my profile")

	// create board and post
	bid := boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
}

func main() {
	println(boards.Render("test_board"))
}

// Output:
// \[[post](/r/demo/modboards?help&__func=CreateThread&bid=1&body.type=textarea)]
//
// ----------------------------------------
// ## [First Post (title)](/r/demo/modboards:test_board/1)
//
// Body of the first post. (body)
// \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm UTC](/r/demo/modboards:test_board/1) \[[x](/r/demo/modboards?help&__func=DeletePost&bid=1&threadid=1&postid=1)] (0 replies)
