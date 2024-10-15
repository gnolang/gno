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
	bid2 := boards.CreateBoard("test_board2")

	// create a repost to a non-existing board
	rid := boards.CreateRepost(5, pid, "", "Check this out", bid2)
	println(rid)
	println(boards.Render("test_board1"))
}

// Error:
// src board not exist
