// PKGPATH: gno.land/r/archive/boards/filetests/z_9_a_filetest

package z_9_a_filetest

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var dstBoard boards.BoardID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)

	dstBoard = boards.CreateBoard(cross2(cur), "dst_board")
}

func main(cur realm) {
	boards.CreateRepost(cross2(cur), 0, 0, "First Post in (title)", "Body of the first post. (body)", dstBoard)
}

// Error:
// invalid non-user call
