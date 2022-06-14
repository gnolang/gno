// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200gnot

import (
	"strconv"

	"gno.land/r/boards"
	"gno.land/r/users"
)

var dstBoard boards.BoardID

func init() {
	users.Register("", "gnouser", "my profile")

	dstBoard = boards.CreateBoard("dst_board")

	boards.CreateRepost(0, 0, "First Post in (title)", "Body of the first post. (body)", dstBoard)
}

func main() {
}

// Error:
// src board not exist