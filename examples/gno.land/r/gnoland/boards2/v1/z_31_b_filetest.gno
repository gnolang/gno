package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	title          = "Foo"
	body           = "bar"
	reply1         = "reply1"
	reply2         = "reply2"
)

var (
	bid  boards2.BoardID
	tid  boards2.PostID
	rid1 boards2.PostID
	rid2 boards2.PostID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	tid = boards2.CreateThread(cross, bid, title, body)
	rid1 = boards2.CreateReply(cross, bid, tid, 0, reply1)
	rid2 = boards2.CreateReply(cross, bid, tid, 0, reply2)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Just get the total count of replies
	total, _ := boards2.GetPosts(bid, tid, 0, 0, 0)
	println(total == 2)

	// Get all replies
	total, replies := boards2.GetPosts(bid, tid, 0, 0, 10)
	println(len(replies) == 2)

	// Check contents
	println(replies[0].ID == rid1)
	println(replies[0].Creator == owner)
	println(replies[0].Body == reply1)

	println(replies[1].ID == rid2)
	println(replies[1].Creator == owner)
	println(replies[1].Body == reply2)

	// Get starting from the second reply
	total, replies = boards2.GetPosts(bid, tid, 0, 1, 10)
	println(len(replies) == 1)
	println(replies[0].ID == rid2)
}

// Output:
// true
// true
// true
// true
// true
// true
// true
// true
// true
// true
