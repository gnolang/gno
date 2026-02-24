package main

// SEND: 1000000ugnot

import (
	"strconv"
	"testing"

	"gno.land/r/archive/boards"
	susers "gno.land/r/sys/users"
)

var (
	bid boards.BoardID
	pid boards.PostID
	rid boards.PostID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))

	bid = boards.CreateBoard(cross, "test_board")
	pid = boards.CreateThread(cross, bid, "First Post in (title)", "Body of the first post. (body)")
	rid = boards.CreateReply(cross, bid, pid, pid, "First reply of the First post\n")
}

func main() {
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	boards.EditPost(cross, bid, pid, rid, "", "Edited: First reply of the First post\n")
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
// > Edited: First reply of the First post
// >
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/1/2) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=2&threadid=1)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=2&threadid=1)]
//
