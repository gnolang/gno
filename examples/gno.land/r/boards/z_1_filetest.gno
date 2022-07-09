// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200ugnot

import (
	"gno.land/r/boards"
	"gno.land/r/users"
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
//  * [/r/boards:test_board_1](/r/boards:test_board_1)
//  * [/r/boards:test_board_2](/r/boards:test_board_2)
