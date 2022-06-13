// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 2000gnot

import (
	"strconv"

	"gno.land/r/boards"
	"gno.land/r/users"
)

var bid boards.BoardID
var pid boards.PostID
var rid boards.PostID

func init() {
	users.Register("", "gnouser", "my profile")

	bid = boards.CreateBoard("test_board")
	pid = boards.CreateThread(bid, "First Post in (title)", "Body of the first post. (body)")
	rid = boards.CreateReply(bid, pid, pid, "First reply of the First post\n")
}

func main() {
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
	boards.EditPost(bid, pid, rid, "", "Edited: First reply of the First post\n")
	println("----------------------------------------------------")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/1) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=1&postid=1&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)]
//
// > First reply of the First post
// >
// > \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/1/2) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=1&postid=2&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=1&postid=2)]
//
// ----------------------------------------------------
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/1) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=1&postid=1&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)]
//
// > Edited: First reply of the First post
// >
// > \- [@gnouser](/r/users:gnouser), [1970-01-01 12:00am (UTC)](/r/boards:test_board/1/2) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=1&postid=2&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=1&postid=2)]