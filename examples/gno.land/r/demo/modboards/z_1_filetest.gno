// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"gno.land/r/demo/modboards"
	"gno.land/r/demo/users"
)

var board *boards.Board

func init() {
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
//  * [/r/demo/modboards:test_board_1](/r/demo/modboards:test_board_1)
//  * [/r/demo/modboards:test_board_2](/r/demo/modboards:test_board_2)
