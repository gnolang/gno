// PKGPATH: gno.land/r/archive/boards/filetests/z_1_filetest

package z_1_filetest

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var board *boards.Board

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)

	_ = boards.CreateBoard(cross2(cur), "test_board_1")
	_ = boards.CreateBoard(cross2(cur), "test_board_2")
}

func main() {
	println(boards.Render(""))
}

// Output:
// These are all the boards of this realm:
//
//  * [/r/archive/boards:test_board_1](/r/archive/boards:test_board_1)
//  * [/r/archive/boards:test_board_2](/r/archive/boards:test_board_2)
//
