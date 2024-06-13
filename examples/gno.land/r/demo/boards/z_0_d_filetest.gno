// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var bid boards.BoardID

func init() {
	users.Register("", "gnouser", "my profile")
	bid = boards.CreateBoard("test_board")
	boards.CreateReply(bid, 0, 0, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// thread not exist
