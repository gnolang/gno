package main

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

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	uinit.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))

	bid = boards.CreateBoard(cross, "test_board")
	boards.CreateThread(cross, bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreateThread(cross, bid, "Second Post (title)", "Body of the second post. (body)")
	rid := boards.CreateReply(cross, bid, pid, pid, "Reply of the second post")
	println(rid)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	rid2 := boards.CreateReply(cross, bid, pid, pid, "Second reply of the second post")
	println(rid2)
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// 3
// 4
// # Second Post (title)
//
// Body of the second post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/2) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=2&threadid=2)] \[[repost](/r/archive/boards$help&func=CreateRepost&bid=1&postid=2)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=2&threadid=2)]
//
// > Reply of the second post
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/2/3) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=3&threadid=2)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=3&threadid=2)]
//
// > Second reply of the second post
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/2/4) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=4&threadid=2)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=4&threadid=2)]
//
