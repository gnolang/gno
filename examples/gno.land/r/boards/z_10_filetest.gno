// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"strconv"

	"gno.land/r/boards"
	"gno.land/r/users"
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
	boards.DeletePost(bid, pid, pid, "")
	println("----------------------------------------------------")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/1) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=1&postid=1&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)]
//
// ----------------------------------------------------
// thread does not exist with id: 1
