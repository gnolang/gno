// PKGPATH: gno.land/r/archive/boards/filetests/z_12_filetest

package z_12_filetest

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var (
	bid1 boards.BoardID
	bid2 boards.BoardID
	pid  boards.PostID
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetHeight(0)
	uinit.RegisterUser(cross(cur), "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetHeight(123)

	bid1 = boards.CreateBoard(cross(cur), "test_board1")
	pid = boards.CreateThread(cross(cur), bid1, "First Post (title)", "Body of the first post. (body)")
	bid2 = boards.CreateBoard(cross(cur), "test_board2")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	rid := boards.CreateRepost(cross(cur), bid1, pid, "", "Check this out", bid2)
	println(rid)
	println(boards.Render("test_board2"))
}

// Output:
// 1
// \[[post](/r/archive/boards$help&func=CreateThread&bid=2)]
//
// ----------------------------------------
// Repost: Check this out
// ## [First Post (title)](/r/archive/boards:test_board1/1)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm UTC](/r/archive/boards:test_board1/1) \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)] (0 replies) (1 reposts)
//
//
