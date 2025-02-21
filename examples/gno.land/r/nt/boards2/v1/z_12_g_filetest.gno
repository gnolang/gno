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
	tid = boards2.CreateThread(bid, "Foo", "bar")
	rid = boards2.CreateReply(bid, tid, 0, "body")

	// Flag the reply so it's hidden
	boards2.FlagReply(bid, tid, rid, "reason")
}

func main() {
	boards2.EditReply(bid, tid, rid, "body")
}

// Error:
// reply with ID: 2 was hidden
