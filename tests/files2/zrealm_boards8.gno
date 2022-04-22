// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"strconv"

	"gno.land/r/boards"
)

var bid boards.BoardID
var pid boards.PostID
var rid boards.PostID

func init() {
	bid = boards.CreateBoard("test_board")
	boards.CreatePost(bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreatePost(bid, "Second Post (title)", "Body of the second post. (body)")
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
// \- [g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985](/r/users:g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2/3) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=3&body.type=textarea)] \[[x](/r/boards?help?__func=DeletePost&bid=1&threadid=2&postid=3)]
//
// _[see all 1 replies](/r/boards:test_board/2/3)_
//
// > First reply of the first reply
// >
// > \- [g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985](/r/users:g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985), [1970-01-01 12:00am (UTC)](/r/boards:test_board/2/5) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=2&postid=5&body.type=textarea)] \[[x](/r/boards?help?__func=DeletePost&bid=1&threadid=2&postid=5)]
