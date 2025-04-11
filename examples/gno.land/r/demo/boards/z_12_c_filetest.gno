// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 1000000ugnot

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

func main() {
	testing.SetRealm(runtime.NewUserRealm(chain.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register("gnouser123")
	bid1 := boards.CreateBoard("test_board1")
	boards.CreateThread(bid1, "First Post (title)", "Body of the first post. (body)")
	bid2 := boards.CreateBoard("test_board2")

	// create a repost to a non-existing thread
	rid := boards.CreateRepost(bid1, 5, "", "Check this out", bid2)
	println(rid)
	println(boards.Render("test_board1"))
}

// Error:
// thread not exist
