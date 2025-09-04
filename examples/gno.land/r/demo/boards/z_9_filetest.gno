package main

// SEND: 1000000ugnot

import (
	"chain"
	"chain/runtime"
	"strconv"
	"testing"

	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

var (
	firstBoard  boards.BoardID
	secondBoard boards.BoardID
	pid         boards.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(chain.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")

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
// \- [@gnouser123](/u/gnouser123), [2009-02-13 11:31pm (UTC)](/r/demo/boards:second_board/1/1) \[[reply](/r/demo/boards$help&func=CreateReply&bid=2&postid=1&threadid=1)] \[[x](/r/demo/boards$help&func=DeletePost&bid=2&postid=1&threadid=1)]
//
