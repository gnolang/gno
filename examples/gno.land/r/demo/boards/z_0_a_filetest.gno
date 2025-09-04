// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 1000000ugnot

import (
	"chain/runtime"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
)

var bid boards.BoardID

func init() {
	caller := testutils.TestAddress("caller")
	testing.SetRealm(runtime.NewUserRealm(caller))
	bid = boards.CreateBoard(cross, "test_board")
	boards.CreateThread(cross, bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreateThread(cross, bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(cross, bid, pid, pid, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// unauthorized
