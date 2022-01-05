// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"gno.land/r/boards"
)

var board *boards.Board

func init() {
	_ = boards.CreateBoard("test_board_1")
	_ = boards.CreateBoard("test_board_2")
}

func main() {
	println(boards.Render(""))
}

// Output:
//  * [test_board_1](gno.land/r/boards/test_board_1)
//  * [test_board_2](gno.land/r/boards/test_board_2)
