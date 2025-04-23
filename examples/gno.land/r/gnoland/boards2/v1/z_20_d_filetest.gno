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
	rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test123")
	tid = cross(boards2.CreateThread)(bid, "foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "reply")

	testing.SetOriginCaller(owner)
}

func main() {
	// unfrozen thread - should work
	r2 := cross(boards2.CreateReply)(bid, tid, rid, "child 1")

	// frozen thread - can't reply
	cross(boards2.FreezeReply)(bid, tid, rid)
	cross(boards2.CreateReply)(bid, tid, r2, "child 2")
}

// Error:
// reply is frozen
