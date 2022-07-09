// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 199ugnot

import (
	"gno.land/r/boards"
	"gno.land/r/users"
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
// payment must not be less than 200
