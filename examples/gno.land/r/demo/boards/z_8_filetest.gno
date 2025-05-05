package main

// SEND: 1000000ugnot

import (
	"std"
	"strconv"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	bid boards.BoardID
	pid boards.PostID
	rid boards.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	cross(users.Register)("gnouser123")

	bid = cross(boards.CreateBoard)("test_board")
	cross(boards.CreateThread)(bid, "First Post (title)", "Body of the first post. (body)")
	pid = cross(boards.CreateThread)(bid, "Second Post (title)", "Body of the second post. (body)")
	rid = cross(boards.CreateReply)(bid, pid, pid, "Reply of the second post")
}

func main() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	cross(boards.CreateReply)(bid, pid, pid, "Second reply of the second post\n")
	rid2 := cross(boards.CreateReply)(bid, pid, rid, "First reply of the first reply\n")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid)) + "/" + strconv.Itoa(int(rid2))))
}

// Output:
// _[see thread](/r/demo/boards:test_board/2)_
//
// Reply of the second post
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/3) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=3&threadid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=3&threadid=2)]
//
// _[see all 1 replies](/r/demo/boards:test_board/2/3)_
//
// > First reply of the first reply
// >
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/2/5) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=5&threadid=2)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=5&threadid=2)]
//
