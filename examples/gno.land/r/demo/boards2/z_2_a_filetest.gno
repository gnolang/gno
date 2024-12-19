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
	bid := boards2.CreateBoard("test1")
	pid := boards2.CreateThread(bid, "thread", "thread")
	boards2.FlagThread(bid, pid, "reason")
	rid := boards2.CreateReply(bid, pid, pid, "reply")
}

// Error:
// thread with ID: 1 was hidden
