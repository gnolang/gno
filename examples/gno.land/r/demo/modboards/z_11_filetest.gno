// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 2000000000ugnot

import (
	"strconv"

	"gno.land/r/demo/modboards"
	"gno.land/r/demo/users"
)

var (
	bid boards.BoardID
	pid boards.PostID
)

func init() {
	users.Register("", "gnouser", "my profile")

	bid = boards.CreateBoard("test_board")
	pid = boards.CreateThread(bid, "First Post in (title)", "Body of the first post. (body)")
}

func main() {
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
	boards.EditPost(bid, pid, pid, "Edited: First Post in (title)", "Edited: Body of the first post. (body)")
	println("----------------------------------------------------")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/modboards:test_board/1) \[[reply](/r/demo/modboards?help&__func=CreateReply&bid=1&threadid=1&postid=1&body.type=textarea)] \[[x](/r/demo/modboards?help&__func=DeletePost&bid=1&threadid=1&postid=1)]
//
// ----------------------------------------------------
// # Edited: First Post in (title)
//
// Edited: Body of the first post. (body)
// \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/modboards:test_board/1) \[[reply](/r/demo/modboards?help&__func=CreateReply&bid=1&threadid=1&postid=1&body.type=textarea)] \[[x](/r/demo/modboards?help&__func=DeletePost&bid=1&threadid=1&postid=1)]
