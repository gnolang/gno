// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var dstBoard boards.BoardID

func init() {
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
	users.Register("", "gnouser", "my profile")

	dstBoard = boards.CreateBoard("dst_board")

	boards.CreateRepost(0, 0, "First Post in (title)", "Body of the first post. (body)", dstBoard)
}

func main() {
}

// Error:
// src board not exist
