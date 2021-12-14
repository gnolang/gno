// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"gno.land/r/boards"
)

var board *boards.Board

func init() {
	_ = boards.CreateBoard("Test Board #1")
	_ = boards.CreateBoard("Test Board #2")
}

func main() {
	println(boards.Render(""))
}

// Output:
// ## Test Board #1 ##
// ## Test Board #2 ##
