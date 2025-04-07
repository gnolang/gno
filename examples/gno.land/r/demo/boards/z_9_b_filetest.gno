// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	srcBoard boards.BoardID
	pid      boards.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register("gnouser123")

	srcBoard = boards.CreateBoard("first_board")
	pid = boards.CreateThread(srcBoard, "First Post in (title)", "Body of the first post. (body)")

	boards.CreateRepost(srcBoard, pid, "First Post in (title)", "Body of the first post. (body)", 0)
}

func main() {
}

// Error:
// panic: dst board not exist
