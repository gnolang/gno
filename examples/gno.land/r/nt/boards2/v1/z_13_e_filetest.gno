package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid      boards2.BoardID
	rid, tid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar", false)
	rid = boards2.CreateReply(bid, tid, 0, "body")
	boards2.FlagReply(bid, tid, rid, "")
}

func main() {
	boards2.FlagReply(bid, tid, rid, "")
}

// Error:
// item flag count threshold exceeded: 1
