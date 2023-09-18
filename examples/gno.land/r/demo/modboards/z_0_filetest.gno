// PKGPATH: gno.land/r/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"gno.land/r/demo/modboards"
	"gno.land/r/demo/users"
)

var bid boards.BoardID

func init() {
	users.Register("", "gnouser", "my profile")

	bid = boards.CreateBoard("test_board")
	boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	pid := boards.CreateThread(bid, "Second Post (title)", "Body of the second post. (body)")
	boards.CreateReply(bid, pid, pid, "Reply of the second post")
}

func main() {
	println(boards.Render("test_board"))
}

// Output:
// \[[post](/r/demo/modboards?help&__func=CreateThread&bid=1&body.type=textarea)]\[[see flags](/r/demo/modboards:test_board/flags)]
//
// ----------------------------------------
// ## [First Post (title)](/r/demo/modboards:test_board/1)
//
// Body of the first post. (body)
// \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm UTC](/r/demo/modboards:test_board/1) \[[🚩](/r/demo/modboards?help&__func=FlagPost&boardID=1&threadID=1&postID=1)] \[[x](/r/demo/modboards?help&__func=DeletePost&bid=1&threadid=1&postid=1)] (0 replies)
//
// ----------------------------------------
// ## [Second Post (title)](/r/demo/modboards:test_board/2)
//
// Body of the second post. (body)
// \- [@gnouser](/r/users:gnouser), [2009-02-13 11:31pm UTC](/r/demo/modboards:test_board/2) \[[🚩](/r/demo/modboards?help&__func=FlagPost&boardID=1&threadID=2&postID=2)] \[[x](/r/demo/modboards?help&__func=DeletePost&bid=1&threadid=2&postid=2)] (1 replies)
