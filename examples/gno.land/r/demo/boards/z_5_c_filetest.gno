package main

// SEND: 2000000000ugnot

import (
	"std"
	"strconv"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/boards"
	"gno.land/r/demo/users"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	users.Register("", "gnouser", "my profile")
	// create board via registered user
	bid := boards.CreateBoard("test_board")

	// create post via anon user
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test2)
	std.TestSetOrigSend(std.Coins{{"ugnot", 101000000}}, nil)

	pid := boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	boards.CreateReply(bid, pid, pid, "Reply of the first post")

	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post (title)
//
// Body of the first post. (body)
// \- [g1w3jhxapjta047h6lta047h6lta047h6laqcyu4](/r/demo/users:g1w3jhxapjta047h6lta047h6lta047h6laqcyu4), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1) \[[reply](/r/demo/boards?help&__func=CreateReply&bid=1&threadid=1&postid=1&body.type=textarea)] \[[repost](/r/demo/boards?help&__func=CreateRepost&bid=1&postid=1&title.type=textarea&body.type=textarea&dstBoardID.type=textarea)] \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)]
//
// > Reply of the first post
// > \- [g1w3jhxapjta047h6lta047h6lta047h6laqcyu4](/r/demo/users:g1w3jhxapjta047h6lta047h6lta047h6laqcyu4), [2009-02-13 11:31pm (UTC)](/r/demo/boards:test_board/1/2) \[[reply](/r/demo/boards?help&__func=CreateReply&bid=1&threadid=1&postid=2&body.type=textarea)] \[[x](/r/demo/boards?help&__func=DeletePost&bid=1&threadid=1&postid=2)]
