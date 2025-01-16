// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"std"
	"strconv"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	bid boards.BoardID
	pid boards.PostID
)

func init() {
	std.TestSetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register("gnouser123")

	bid = boards.CreateBoard("test_board")
	pid = boards.CreateThread(bid, "First Post in (title)", "Body of the first post. (body)")
}

func main() {
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
	// pid of 2 not exist
	boards.DeletePost(bid, 2, 2, "")
	println("----------------------------------------------------")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Error:
// thread not exist
