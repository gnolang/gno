// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	pboards "gno.land/p/demo/boards"
	"gno.land/r/demo/boards"
)

var bid pboards.BoardID

func init() {
	bid = boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(bid, pid, pid, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// unauthorized
