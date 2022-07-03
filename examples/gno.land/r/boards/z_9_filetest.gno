// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200ugnot

import (
	"strconv"

	"gno.land/r/boards"
	"gno.land/r/users"
)

var (
	firstBoard  boards.BoardID
	secondBoard boards.BoardID
	pid         boards.PostID
)

func init() {
	users.Register("", "gnouser", "my profile")

	firstBoard = boards.CreateBoard("first_board")
	secondBoard = boards.CreateBoard("second_board")
	pid = boards.CreateThread(firstBoard, "First Post in (title)", "Body of the first post. (body)")

	boards.CreateRepost(firstBoard, pid, "First Post in (title)", "Body of the first post. (body)", secondBoard)
}

func main() {
	println(boards.Render("second_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:second_board/1/1) \[[reply](/r/boards?help&__func=CreateReply&bid=2&threadid=1&postid=1&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=2&threadid=1&postid=1)]
