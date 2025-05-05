// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 1000000ugnot

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
)

var bid boards.BoardID

func init() {
	caller := testutils.TestAddress("caller")
	testing.SetRealm(std.NewUserRealm(caller))
	bid = cross(boards.CreateBoard)("test_board")
	cross(boards.CreateThread)(bid, "First Post (title)", "Body of the first post. (body)")
	pid := cross(boards.CreateThread)(bid, "Second Post (title)", "Body of the second post. (body)")
	cross(boards.CreateReply)(bid, pid, pid, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// unauthorized
