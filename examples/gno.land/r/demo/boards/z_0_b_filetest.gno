// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 199000000ugnot

import (
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var bid boards.BoardID

func init() {
	users.Register("", "gnouser", "my profile")
	bid = boards.CreateBoard("test_board")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// payment must not be less than 200000000
