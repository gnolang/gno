package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	comment = "Second comment"
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board", false)
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar", false)
	rid = cross(boards2.CreateReply)(bid, tid, 0, "Parent comment")

	// Flag parent post so it's hidden
	cross(boards2.FlagReply)(bid, tid, rid, "")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.CreateReply)(bid, tid, rid, "Sub comment")
}

// Error:
// replying to a hidden or frozen reply is not allowed
