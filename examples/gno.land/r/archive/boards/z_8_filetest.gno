package main

// SEND: 1000000ugnot

import (
	"strconv"
	"testing"

	"gno.land/r/archive/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	bid boards.BoardID
	pid boards.PostID
	rid boards.PostID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")

	bid = boards.CreateBoard(cross, "test_board")
	boards.CreateThread(cross, bid, "First Post (title)", "Body of the first post. (body)")
	pid = boards.CreateThread(cross, bid, "Second Post (title)", "Body of the second post. (body)")
	rid = boards.CreateReply(cross, bid, pid, pid, "Reply of the second post")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	boards.CreateReply(cross, bid, pid, pid, "Second reply of the second post\n")
	rid2 := boards.CreateReply(cross, bid, pid, rid, "First reply of the first reply\n")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid)) + "/" + strconv.Itoa(int(rid2))))
}

// Output:
// _[see thread](/r/archive/boards:test_board/2)_
//
// Reply of the second post
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/2/3) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=3&threadid=2)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=3&threadid=2)]
//
// _[see all 1 replies](/r/archive/boards:test_board/2/3)_
//
// > First reply of the first reply
// >
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/2/5) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=5&threadid=2)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=5&threadid=2)]
//
