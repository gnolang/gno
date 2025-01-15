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
	// Only single user per flag can't be tested atm, as flagThreshold = 1.
	bid := boards2.CreateBoard("test1")
	pid := boards2.CreateThread(bid, "thread", "thread")

	boards2.FlagThread(bid, pid, "reason1")
	boards2.FlagThread(bid, pid, "reason2")
}

// Error:
// item flag count threshold exceeded: 1
