// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"strconv"

	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
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
	rid := boards.CreateReply(bid, pid, pid, "Reply of the second post")
}

func main() {
	rid2 := boards.CreateReply(bid, pid, pid, "Second reply of the second post\n")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # Second Post (title)
//
// Body of the second post. (body)
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=2&postid=2)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=2&postid=2)]
//
// > Reply of the second post
// > \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/3) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=2&postid=3)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=2&postid=3)]
//
// > Second reply of the second post
// >
// > \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/4) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=2&postid=4)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=2&postid=4)]