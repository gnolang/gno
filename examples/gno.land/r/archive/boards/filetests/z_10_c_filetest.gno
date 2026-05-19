// PKGPATH: gno.land/r/archive/boards/filetests/z_10_c_filetest

package z_10_c_filetest

// SEND: 1000000ugnot

import (
	"strconv"
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var (
	bid boards.BoardID
	pid boards.PostID
	rid boards.PostID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)

	bid = boards.CreateBoard(cross(cur), "test_board")
	pid = boards.CreateThread(cross(cur), bid, "First Post in (title)", "Body of the first post. (body)")
	rid = boards.CreateReply(cross(cur), bid, pid, pid, "First reply of the First post\n")
}

func main(cur realm) {
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	boards.DeletePost(cross(cur), bid, pid, rid, "")
	println("----------------------------------------------------")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/1) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=1&threadid=1)] \[[repost](/r/archive/boards$help&func=CreateRepost&bid=1&postid=1)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)]
//
// > First reply of the First post
// >
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/1/2) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=2&threadid=1)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=2&threadid=1)]
//
// ----------------------------------------------------
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/1) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=1&threadid=1)] \[[repost](/r/archive/boards$help&func=CreateRepost&bid=1&postid=1)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)]
//
