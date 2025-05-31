package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid boards2.BoardID
	tid boards2.PostID
	rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test123", false)
	tid = cross(boards2.CreateThread)(bid, "foo", "bar", false)
	rid = cross(boards2.CreateReply)(bid, tid, 0, "reply")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	// unfrozen thread - should work
	cross(boards2.CreateReply)(bid, tid, rid, "child 1")

	// frozen thread - can't reply
	cross(boards2.FreezeReply)(bid, tid, rid)
	cross(boards2.CreateReply)(bid, tid, rid, "child 2")
}

// Error:
// replying to a hidden or frozen reply is not allowed
