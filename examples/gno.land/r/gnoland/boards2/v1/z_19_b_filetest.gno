package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	tid boards2.PostID
)

func init() {
	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")
	tid = boards2.CreateThread(bid, "foo", "bar")

	boards2.FreezeThread(bid, tid)
}

func main() {
	// Attempt to freeze a frozen thread
	boards2.FreezeThread(bid, tid)
}

// Error:
// thread is frozen
