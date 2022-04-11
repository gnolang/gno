// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"strconv"

	"gno.land/r/boards"
)

var bid boards.BoardID
var pid boards.PostID

func init() {
	bid = boards.CreateBoard("test_board")
	boards.CreatePost(bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreatePost(bid, "Second Post (title)", "Body of the second post. (body)")
	rid := boards.CreateReply(bid, pid, "Reply of the second post")
}

func main() {
	rid2 := boards.CreateReply(bid, pid, "Second reply of the second post\n")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # Second Post (title)
//
// Body of the second post. (body)
// - by g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2) [reply](/r/boards?help&__func=CreateReply&bid=1&postid=2&body.type=textarea)
//
// > Reply of the second post
// > - by g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2#3) [reply](/r/boards?help&__func=CreateReply&bid=1&postid=3&body.type=textarea)
//
// > Second reply of the second post
// >
// > - by g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985, [1970-01-01 12:00am (UTC)](/r/boards:test_board/2#4) [reply](/r/boards?help&__func=CreateReply&bid=1&postid=4&body.type=textarea)
