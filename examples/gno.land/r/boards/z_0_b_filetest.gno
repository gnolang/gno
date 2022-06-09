// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 199gnot

import (
	"gno.land/r/boards"
	"gno.land/r/users"
)

var bid boards.BoardID

func init() {
	users.Register("", "gnouser", "my profile")

	bid = boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(bid, pid, pid, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Error:
// payment must not be less than 200