// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"std"
	"strconv"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	bid boards.BoardID
	pid boards.PostID
)

func init() {
	std.TestSetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register("gnouser123")

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
// \- [@gnouser123](/r/gnoland/users/v1:gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=2&postid=2)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=2&postid=2)]
//
// > Reply of the second post
// > \- [@gnouser123](/r/gnoland/users/v1:gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/3) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=2&postid=3)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=2&postid=3)]
//
