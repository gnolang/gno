// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"std"
	"strconv"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var dstBoard boards.BoardID

func init() {
	std.TestSetOrigCaller(testutils.TestAddress("test"))
	users.Register("", "gnouser", "my profile")

	dstBoard = boards.CreateBoard("dst_board")

	boards.CreateRepost(0, 0, "First Post in (title)", "Body of the first post. (body)", dstBoard)
}

func main() {
}

// Error:
// src board not exist
