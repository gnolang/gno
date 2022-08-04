// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 2000000000ugnot

import (
	"strconv"

	"gno.land/r/boards"
	"gno.land/r/users"
)

var (
	bid boards.BoardID
	pid boards.PostID
	rid boards.PostID
)

func init() {
	users.Register("", "gnouser", "my profile")

	bid = boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
	rid = boards.CreateReply(bid, pid, pid, "Reply of the second post")
}

func main() {
	boards.CreateReply(bid, pid, pid, "Second reply of the second post\n")
	rid2 := boards.CreateReply(bid, pid, rid, "First reply of the first reply\n")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid)) + "/" + strconv.Itoa(int(rid2))))
}

// Output:
// _[see thread](/r/boards:test_board/2)_
//
// Reply of the second post
// \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/boards:test_board/2/3) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=3&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=3)]
//
// _[see all 1 replies](/r/boards:test_board/2/3)_
//
// > First reply of the first reply
// >
// > \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/boards:test_board/2/5) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=5&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=5)]
