package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid      boards2.BoardID
	pid, rid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)

	bid = boards2.CreateBoard("test123")
	pid = boards2.CreateThread(bid, "thread", "thread")
	rid = boards2.CreateReply(bid, pid, pid, "reply1")
}

func main() {
	boards2.FlagReply(bid, pid, rid, "reason")

	// Ensure that nested replies are denied if parent reply is hidden
	boards2.CreateReply(bid, pid, rid, "reply1.1")
}

// Error:
// reply with ID: 2 was hidden
