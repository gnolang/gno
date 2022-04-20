// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 2000gnot

import (
	"gno.land/r/boards"
	"gno.land/r/users"
)

func init() {
	// register
	users.Register("", "gnouser", "my profile")

	// create board and post
	bid := boards.CreateBoard("test_board")
	boards.CreatePost(bid, "First Post (title)", "Body of the first post. (body)")
}

func main() {
	println(boards.Render("test_board"))
}

// Output:
// \[[post](/r/boards?help&__func=CreatePost&bid=1&body.type=textarea)]
//
// ----------------------------------------
// ## [First Post (title)](/r/boards:test_board/1)
//
// Body of the first post. (body)
// \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am UTC](/r/boards:test_board/1) (0 replies)
