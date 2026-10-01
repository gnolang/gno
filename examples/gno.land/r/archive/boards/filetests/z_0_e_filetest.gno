// PKGPATH: gno.land/r/archive/boards/filetests/z_0_e
package z_0_e

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var bid boards.BoardID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)
	boards.CreateReply(cross(cur), bid, 0, 0, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// board not exist
