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
	pid = cross(boards.CreateThread)(bid, "First Post in (title)", "Body of the first post. (body)")
	rid = cross(boards.CreateReply)(bid, pid, pid, "First reply of the First post\n")
}

func main() {
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")))
	cross(boards.EditPost)(bid, pid, rid, "", "Edited: First reply of the First post\n")
	println("----------------------------------------------------")
	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=1&threadid=1)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=1)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)]
//
// > First reply of the First post
// > 
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1/2) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=2&threadid=1)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=2&threadid=1)]
//
// ----------------------------------------------------
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=1&threadid=1)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=1)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)]
//
// > Edited: First reply of the First post
// > 
// > \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1/2) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=2&threadid=1)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=2&threadid=1)]
//
