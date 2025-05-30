// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var bid boards.BoardID

func init() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	cross(users.Register)("gnouser123")
	cross(boards.CreateReply)(bid, 0, 0, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// board not exist
