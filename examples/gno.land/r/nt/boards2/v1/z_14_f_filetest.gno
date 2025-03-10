package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar", false)
	rid = boards2.CreateReply(bid, tid, 0, "body")

	// Call using a user that has not permission to delete replies
	std.TestSetOriginCaller(user)
}

func main() {
	boards2.DeleteReply(bid, tid, rid)
}

// Error:
// unauthorized
