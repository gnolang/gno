package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner         address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	title                 = "Foo"
	body                  = "bar"
	reply                 = "reply"
	reply1ToReply         = "reply1 to reply"
	reply2ToReply         = "reply2 to reply"
)

var (
	bid         boards2.BoardID
	tid         boards2.PostID
	rid         boards2.PostID
	rid1ToReply boards2.PostID
	rid2ToReply boards2.PostID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	tid = boards2.CreateThread(cross, bid, title, body)
	rid = boards2.CreateReply(cross, bid, tid, 0, reply)
	rid1ToReply = boards2.CreateReply(cross, bid, tid, rid, reply1ToReply)
	rid2ToReply = boards2.CreateReply(cross, bid, tid, rid, reply2ToReply)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Just get the total count of replies to reply
	total, _ := boards2.GetPosts(bid, tid, rid, 0, 0)
	println(total == 2)

	// Get all replies to reply
	total, replies := boards2.GetPosts(bid, tid, rid, 0, 10)
	println(len(replies) == 2)

	// Check contents
	println(replies[0].ID == rid1ToReply)
	println(replies[0].Creator == owner)
	println(replies[0].Body == reply1ToReply)

	println(replies[1].ID == rid2ToReply)
	println(replies[1].Creator == owner)
	println(replies[1].Body == reply2ToReply)

	// Get starting from the second reply to reply
	total, replies = boards2.GetPosts(bid, tid, rid, 1, 10)
	println(len(replies) == 1)
	println(replies[0].ID == rid2ToReply)
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
