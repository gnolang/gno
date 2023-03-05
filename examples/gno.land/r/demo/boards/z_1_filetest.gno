// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var board *boards.Board

func init() {
	std.TestSetOrigCaller(testutils.TestAddress("test"))
	users.Register("", "gnouser", "my profile")

	_ = boards.CreateBoard("test_board_1")
	_ = boards.CreateBoard("test_board_2")
}

func main() {
	println(boards.Render(""))
}

// Output:
// These are all the boards of this realm:
//
//  * [/r/demo/boards:test_board_1](/r/demo/boards:test_board_1)
//  * [/r/demo/boards:test_board_2](/r/demo/boards:test_board_2)
