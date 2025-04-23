package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test123")
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "reply1")

	// Hide thread by flagging it so reply of a reply can't be submitted
	cross(boards2.FlagThread)(bid, tid, "reason")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.CreateReply)(bid, tid, rid, "reply1.1")
}

// Error:
// thread is hidden
