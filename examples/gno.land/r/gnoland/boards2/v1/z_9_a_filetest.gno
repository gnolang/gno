package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	title = "Test Thread"
	body  = "Test body"
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	pid = boards2.CreateThread(bid, title, body)
}

func main() {
	boards2.DeleteThread(bid, pid)

	// Ensure thread doesn't exist
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread does not exist with ID: 1
