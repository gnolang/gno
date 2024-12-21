// PKGPATH: gno.land/r/demo/boards_test
package boards_test

// SEND: 200000000ugnot

import (
	"std"
	"strconv"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

var (
	bid boards.BoardID
	pid boards.PostID
	rid boards.PostID
)

func init() {
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
	users.Register("", "gnouser", "my profile")

	bid = boards.CreateBoard("test_board")
	pid = boards.CreateThread(bid, "First Post in (title)", "Body of the first post. (body)")
	rid = boards.CreateReply(bid, pid, pid, "First reply of the First post\n")
}

func main() {
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
	boards.DeletePost(bid, pid, rid, "")
	println("----------------------------------------------------")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=1&postid=1)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=1)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=1&postid=1)]
//
// > First reply of the First post
// >
// > \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1/2) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=1&postid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=1&postid=2)]
//
// ----------------------------------------------------
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser](/r/demo/users:gnouser), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&threadid=1&postid=1)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=1)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&threadid=1&postid=1)]
//
