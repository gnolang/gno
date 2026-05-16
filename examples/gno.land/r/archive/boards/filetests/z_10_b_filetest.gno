// PKGPATH: gno.land/r/archive/boards/filetests/z_10_b_filetest

package z_10_b_filetest

// SEND: 1000000ugnot

import (
	"strconv"
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var (
	bid boards.BoardID
	pid boards.PostID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross2(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)

	bid = boards.CreateBoard(cross2(cur), "test_board")
	pid = boards.CreateThread(cross2(cur), bid, "First Post in (title)", "Body of the first post. (body)")
}

func main(cur realm) {
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
	// pid of 2 not exist
	boards.DeletePost(cross2(cur), bid, 2, 2, "")
	println("----------------------------------------------------")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Error:
// invalid non-user call
