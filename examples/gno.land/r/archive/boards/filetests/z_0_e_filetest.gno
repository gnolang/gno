// PKGPATH: gno.land/r/archive/boards_test
package boards_test

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var bid boards.BoardID

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	uinit.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	boards.CreateReply(cross, bid, 0, 0, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// board not exist
