// PKGPATH: gno.land/r/archive/boards/filetests/z_9_b_filetest

package z_9_b_filetest

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var (
	srcBoard boards.BoardID
	pid      boards.PostID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)

	srcBoard = boards.CreateBoard(cross2(cur), "first_board")
	pid = boards.CreateThread(cross2(cur), srcBoard, "First Post in (title)", "Body of the first post. (body)")
}

func main(cur realm) {
	boards.CreateRepost(cross2(cur), srcBoard, pid, "First Post in (title)", "Body of the first post. (body)", 0)
}

// Error:
// invalid non-user call
