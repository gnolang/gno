// PKGPATH: gno.land/r/boards_test
package boards_test

import (
	"gno.land/r/boards"
)

var bid boards.BoardID

func init() {
	bid = boards.CreateBoard("test_board")
	boards.CreatePost(bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreatePost(bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(bid, pid, pid, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Output:
// \[[post](/r/boards?help&__func=CreatePost&bid=1&body.type=textarea)]
//
// ----------------------------------------
// ## [First Post (title)](/r/boards:test_board/1)
//
// Body of the first post. (body)
// \- [g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985](/r/users:g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985), [1970-01-01 12:00am UTC](/r/boards:test_board/1) \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)] (0 replies)
//
// ----------------------------------------
// ## [Second Post (title)](/r/boards:test_board/2)
//
// Body of the second post. (body)
// \- [g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985](/r/users:g1arjyc64rpthwn8zhxtzjvearm5scy43y7vm985), [1970-01-01 12:00am UTC](/r/boards:test_board/2) \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=2&postid=2)] (1 replies)
