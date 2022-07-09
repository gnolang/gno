// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200ugnot

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
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
}

func main() {
	rid := boards.CreateReply(bid, pid, pid, "Reply of the second post")
	println(rid)
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// 3
// # Second Post (title)
//
// Body of the second post. (body)
// \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=2&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=2)]
//
// > Reply of the second post
// > \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2/3) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=3&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=3)]
