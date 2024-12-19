package main

import (
	"std"

	"gno.land/r/demo/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOrigCaller(owner)
}

func main() {
	// ensure that nested replies denied if root thread is hidden.
	bid := boards2.CreateBoard("test1")
	pid := boards2.CreateThread(bid, "thread", "thread")
	rid := boards2.CreateReply(bid, pid, pid, "reply1")

	boards2.FlagReply(bid, pid, rid, "reason")
	_ = boards2.CreateReply(bid, pid, rid, "reply1.1")
}

// Error:
// thread with ID: 2 was hidden
