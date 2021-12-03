// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"gno.land/p/testutils"
	"gno.land/r/boards"
)

var board *boards.Board

func init() {
	addr0 := testutils.TestAddress("addr0")
	board = boards.NewPrivateBoard("Test Board", addr0)
	board.AddPost(addr0, "First Post (title)", "Body of the first post. (body)")
}

func main() {
	println(board.Render(""))
}

// Output:
// ### Test Board ###
// ----------------------------------------
// # First Post (title)
//
// Body of the first post. (body)
