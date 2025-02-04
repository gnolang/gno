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
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test-board")
	pid = boards2.CreateThread(bid, "Foo", "bar")
}

func main() {
	boards2.FlagThread(bid, pid, "")

	// Ensure that original thread content not visible
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread with ID: 1 has been flagged as inappropriate
