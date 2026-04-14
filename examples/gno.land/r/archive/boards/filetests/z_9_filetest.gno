package main

// SEND: 1000000ugnot

import (
	"strconv"
	"testing"

	"gno.land/r/archive/boards"
	uinit "gno.land/r/sys/users/init"
)

var (
	firstBoard  boards.BoardID
	secondBoard boards.BoardID
	pid         boards.PostID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	uinit.RegisterUser(cross, "gnouser123", address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))

	firstBoard = boards.CreateBoard(cross, "first_board")
	secondBoard = boards.CreateBoard(cross, "second_board")
	pid = boards.CreateThread(cross, firstBoard, "First Post in (title)", "Body of the first post. (body)")

	boards.CreateRepost(cross, firstBoard, pid, "First Post in (title)", "Body of the first post. (body)", secondBoard)
}

func main() {
	println(boards.Render("second_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post in (title)
//
// Body of the first post. (body)
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/archive/boards:second_board/1/1) \[[reply](/r/archive/boards$help&func=CreateReply&bid=2&postid=1&threadid=1)] \[[x](/r/archive/boards$help&func=DeletePost&bid=2&postid=1&threadid=1)]
//
