package main

// SEND: 1000000ugnot

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	users "gno.land/r/gnoland/users/v1"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	testing.SetRealm(std.NewUserRealm(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	cross(users.Register)("gnouser123")
	// create board via registered user
	bid := cross(boards.CreateBoard)("test_board")

	// create post via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(std.Coins{{"ugnot", 101000000}})

	pid := cross(boards.CreateThread)(bid, "First Post (title)", "Body of the first post. (body)")
	cross(boards.CreateReply)(bid, pid, pid, "Reply of the first post")

	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post (title)
//
// Body of the first post. (body)
// \- [g1w3jhxapjta047h6lta047h6lta047h6laqcyu4](/u/g1w3jhxapjta047h6lta047h6lta047h6laqcyu4), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=1&threadid=1)] \[[repost](/r/demo/boards$help&func=CreateRepost&bid=1&postid=1)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)]
//
// > Reply of the first post
// > \- [g1w3jhxapjta047h6lta047h6lta047h6laqcyu4](/u/g1w3jhxapjta047h6lta047h6lta047h6laqcyu4), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1/2) \[[reply](/r/demo/boards$help&func=CreateReply&bid=1&postid=2&threadid=1)] \[[x](/r/demo/boards$help&func=DeletePost&bid=1&postid=2&threadid=1)]
//
