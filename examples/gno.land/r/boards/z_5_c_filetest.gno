package main

// SEND: 2000gnot

import (
	"std"

	"gno.land/p/testutils"
	"gno.land/r/users"
	"gno.land/r/boards"
	"strconv"
)

const admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

func main() {
	users.Register("", "gnouser", "my profile")
	//create board via registered user
	bid := boards.CreateBoard("test_board")

	//create post via anon user
	test2 := testutils.TestAddress("test2")
	std.TestSetOrigCaller(test2)
	std.TestSetOrigSend(std.Coins{{"gnot", 11}}, nil)

	pid := boards.CreateThread(bid, "First Post (title)", "Body of the first post. (body)")
	boards.CreateReply(bid, pid, pid, "Reply of the first post")

	println(boards.Render("test_board/" + strconv.Itoa(int(pid))))
}

// Output:
// # First Post (title)
//
// Body of the first post. (body)
// \- [g1w3jhxapjta047h6lta047h6lta047h6laqcyu4](/r/users:g1w3jhxapjta047h6lta047h6lta047h6laqcyu4), [1970-01-01 12:00am (UTC)](/r/boards:test_board/1) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=1&postid=1&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=1&postid=1)]
//
// > Reply of the first post
// > \- [g1w3jhxapjta047h6lta047h6lta047h6laqcyu4](/r/users:g1w3jhxapjta047h6lta047h6lta047h6laqcyu4), [1970-01-01 12:00am (UTC)](/r/boards:test_board/1/2) \[[reply](/r/boards?help&__func=CreateReply&bid=1&threadid=1&postid=2&body.type=textarea)] \[[x](/r/boards?help&__func=DeletePost&bid=1&threadid=1&postid=2)]