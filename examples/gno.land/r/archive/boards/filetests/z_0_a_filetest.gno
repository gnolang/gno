// PKGPATH: gno.land/r/archive/boards/filetests/z_0_a
package z_0_a

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/archive/boards"
)

var bid boards.BoardID

func init(cur realm) {
	caller := testutils.TestAddress("caller")
	testing.SetRealm(testing.NewUserRealm(caller))
	bid = boards.CreateBoard(cross2(cur), "test_board")
	boards.CreateThread(cross2(cur), bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreateThread(cross2(cur), bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(cross2(cur), bid, pid, pid, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// unauthorized
