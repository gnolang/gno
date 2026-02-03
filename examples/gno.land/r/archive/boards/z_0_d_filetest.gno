// PKGPATH: gno.land/r/archive/boards_test
package boards_test

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	users "gno.land/r/gnoland/users/v1"
)

var bid boards.BoardID

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")
	bid = boards.CreateBoard(cross, "test_board")
	boards.CreateReply(cross, bid, 0, 0, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// thread not exist
