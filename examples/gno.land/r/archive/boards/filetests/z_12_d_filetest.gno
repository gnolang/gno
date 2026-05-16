// PKGPATH: gno.land/r/archive/boards/filetests/z_12_d_filetest

package z_12_d_filetest

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)
	bid1 := boards.CreateBoard(cross2(cur), "test_board1")
	pid := boards.CreateThread(cross2(cur), bid1, "First Post (title)", "Body of the first post. (body)")
	boards.CreateBoard(cross2(cur), "test_board2")

	// create a repost to a non-existing destination board
	rid := boards.CreateRepost(cross2(cur), bid1, pid, "", "Check this out", 5)
	println(rid)
	println(boards.Render("test_board1"))
}

// Error:
// dst board not exist
