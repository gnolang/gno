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
	tid = cross(boards2.CreateThread)(bid, "foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "reply")

	cross(boards2.FreezeThread)(bid, tid)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Attempt to freeze a reply on frozen thread
	cross(boards2.FreezeReply)(bid, tid, rid)
}

// Error:
// thread is frozen
