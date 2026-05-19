// PKGPATH: gno.land/r/archive/boards/filetests/z_5_filetest

package z_5_filetest

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
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)

	bid = boards.CreateBoard(cross(cur), "test_board")
	boards.CreateThread(cross(cur), bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreateThread(cross(cur), bid, "Second Post (title)", "Body of the second post. (body)")
	_ = boards.CreateReply(cross(cur), bid, pid, pid, "Reply of the second post")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	_ = boards.CreateReply(cross(cur), bid, pid, pid, "Second reply of the second post\n")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # Second Post (title)
//
// Body of the second post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/2) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=2&threadid=2)] \[[repost](/r/archive/boards$help&func=CreateRepost&bid=1&postid=2)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=2&threadid=2)]
//
// > Reply of the second post
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/2/3) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=3&threadid=2)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=3&threadid=2)]
//
// > Second reply of the second post
// >
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/2/4) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=4&threadid=2)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=4&threadid=2)]
//
