package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)

	bid = boards2.CreateBoard("test123")
	pid = boards2.CreateThread(bid, "thread", "thread")
}

func main() {
	boards2.FlagThread(bid, pid, "reason1")

	// Check the default flag threshold = 1
	boards2.FlagThread(bid, pid, "reason2")
}

// Error:
// item flag count threshold exceeded: 1
