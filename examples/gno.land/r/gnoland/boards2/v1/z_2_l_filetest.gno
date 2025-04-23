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
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test123")
	tid = cross(boards2.CreateThread)(bid, "foo", "bar")
	cross(boards2.FreezeThread)(bid, tid)

	testing.SetOriginCaller(owner)
}

func main() {
	// cannot reply to a frozen thread
	cross(boards2.CreateReply)(bid, tid, 0, "foobar")
}

// Error:
// thread is frozen
