// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"gno.land/p/testutils"
	"gno.land/r/boards"
)

var board *boards.Board

func init() {
	board = boards.CreateBoard("test_board")
	addr0 := testutils.TestAddress("addr0")
	board.AddPost(addr0, "First Post (title)", "Body of the first post. (body)")
	post2 := board.AddPost(addr0, "Second Post (title)", "Body of the second post. (body)")
	post2.AddReply(addr0, "Reply of the second post")
}

func main() {
	println(board.Render())
}

// Output:
// ### test_board ###
//
// ----------------------------------------
// # First Post (title)
//
// Body of the first post. (body)
// - by g1v9jxgu3sta047h6lta047h6lta047h6l2czdz2, [1970-01-01 12:00am UTC](/r/boards/test_board/1) (0 replies)
// ----------------------------------------
// # Second Post (title)
//
// Body of the second post. (body)
// - by g1v9jxgu3sta047h6lta047h6lta047h6l2czdz2, [1970-01-01 12:00am UTC](/r/boards/test_board/2) (1 replies)
