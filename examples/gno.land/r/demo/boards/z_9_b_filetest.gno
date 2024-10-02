// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var (
	srcBoard boards.BoardID
	pid      boards.PostID
)

func init() {
	users.Register("", "gnouser", "my profile")

	srcBoard = boards.CreateBoard("first_board")
	pid = boards.CreateThread(srcBoard, "First Post in (title)", "Body of the first post. (body)")

	boards.CreateRepost(srcBoard, pid, "First Post in (title)", "Body of the first post. (body)", 0)
}

func main() {
}

// Error:
// dst board not exist
