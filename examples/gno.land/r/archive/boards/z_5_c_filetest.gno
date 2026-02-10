package main

// SEND: 1000000ugnot

import (
	"chain"
	"strconv"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/archive/boards"
	users "gno.land/r/gnoland/users/v1"
)

const admin = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	testing.SetRealm(testing.NewUserRealm(address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))) // so that CurrentRealm.Addr() matches OrigCaller
	users.Register(cross, "gnouser123")
	// create board via registered user
	bid := boards.CreateBoard(cross, "test_board")

	// create post via anon user
	test2 := testutils.TestAddress("test2")
	testing.SetOriginCaller(test2)
	testing.SetOriginSend(chain.Coins{{"ugnot", 101000000}})

	pid := boards.CreateThread(cross, bid, "First Post (title)", "Body of the first post. (body)")
	boards.CreateReply(cross, bid, pid, pid, "Reply of the first post")

	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post (title)
//
// Body of the first post. (body)
// \- [g1w3jhxapjta047h6lta047h6lta047h6laqcyu4](/u/g1w3jhxapjta047h6lta047h6lta047h6laqcyu4), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/1) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=1&threadid=1)] \[[repost](/r/archive/boards$help&func=CreateRepost&bid=1&postid=1)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=1&threadid=1)]
//
// > Reply of the first post
// > \- [g1w3jhxapjta047h6lta047h6lta047h6laqcyu4](/u/g1w3jhxapjta047h6lta047h6lta047h6laqcyu4), [2009-02-13 11:31pm (UTC)](/r/archive/boards:test_board/1/2) \[[reply](/r/archive/boards$help&func=CreateReply&bid=1&postid=2&threadid=1)] \[[x](/r/archive/boards$help&func=DeletePost&bid=1&postid=2&threadid=1)]
//
