package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	pid = boards2.CreateThread(bid, "Foo", "bar", false)

	boards2.SetFlaggingThreshold(bid, 2)
	boards2.FlagThread(bid, pid, "")
}

func main() {
	boards2.FlagThread(bid, pid, "")
}

// Error:
// item has been already flagged by a current user
