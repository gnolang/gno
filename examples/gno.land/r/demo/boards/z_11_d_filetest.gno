// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 2000000000ugnot

import (
	"strconv"

	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var (
	bid boards.BoardID
	pid boards.PostID
	rid boards.PostID
)

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
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1) \[[reply](/r/demo/boards?help&__func=CreateReply&bid=1&threadid=1&postid=1&body.type=textarea)] \[[repost](/r/demo/boards?help&__func=CreateRepost&bid=1&postid=1&title.type=textarea&body.type=textarea&dstBoardID.type=textarea)] \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)]
//
// > First reply of the First post
// >
// > \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1/2) \[[reply](/r/demo/boards?help&__func=CreateReply&bid=1&threadid=1&postid=2&body.type=textarea)] \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=1&postid=2)]
//
// ----------------------------------------------------
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1) \[[reply](/r/demo/boards?help&__func=CreateReply&bid=1&threadid=1&postid=1&body.type=textarea)] \[[repost](/r/demo/boards?help&__func=CreateRepost&bid=1&postid=1&title.type=textarea&body.type=textarea&dstBoardID.type=textarea)] \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)]
//
// > Edited: First reply of the First post
// >
// > \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1/2) \[[reply](/r/demo/boards?help&__func=CreateReply&bid=1&threadid=1&postid=2&body.type=textarea)] \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=1&postid=2)]
