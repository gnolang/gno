package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test123")
	tid = cross(boards2.CreateThread)(bid, "thread", "thread")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "reply1")

	// Hide reply by flagging it so sub reply can't be submitted
	cross(boards2.FlagReply)(bid, tid, rid, "reason")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.CreateReply)(bid, tid, rid, "reply1.1")
}

// Error:
// replying to a hidden or frozen reply is not allowed
