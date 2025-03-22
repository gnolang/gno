package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)

	// Created a board with flagging threshold greater than 1
	bid = boards2.CreateBoard("test-board")
	boards2.SetFlaggingThreshold(bid, 2)

	// Create a thread so the realm owner can flag and hide it with a single flag
	pid = boards2.CreateThread(bid, "Foo", "bar")
	// Also freeze board to make sure that realm owner can still flag the thread
	boards2.FreezeBoard(bid)
}

func main() {
	boards2.FlagThread(bid, pid, "")

	// Ensure that original thread content not visible
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread with ID: 1 has been flagged as inappropriate
