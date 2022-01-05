// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"gno.land/p/testutils"
	"gno.land/r/boards"
)

var board *boards.Board

func init() {
	addr0 := testutils.TestAddress("addr0")
	board = boards.NewPrivateBoard("test_board", addr0)
	board.AddPost(addr0, "First Post (title)", "Body of the first post. (body)")
	post2 := board.AddPost(addr0, "Second Post (title)", "Body of the second post. (body)")
	post2.AddReply(addr0, "Reply of the second post")
}

func main() {
	println(board.Render())
}

// Output:
// ### (private) test_board ###
//
// ----------------------------------------
// # First Post (title)
//
// Body of the first post. (body)
//
//                              (0 replies)
// ----------------------------------------
// # Second Post (title)
//
// Body of the second post. (body)
//
//                              (1 replies)
