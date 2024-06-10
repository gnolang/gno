// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

func main() {
	users.Register("", "gnouser", "my profile")
	bid1 := boards.CreateBoard("test_board1")
	pid := boards.CreateThread(bid1, "First Post (title)", "Body of the first post. (body)")
	boards.CreateBoard("test_board2")

	// create a repost to a non-existing destination board
	rid := boards.CreateRepost(bid1, pid, "", "Check this out", 5)
	println(rid)
	println(boards.Render("test_board1"))
}

// Error:
// dst board not exist
