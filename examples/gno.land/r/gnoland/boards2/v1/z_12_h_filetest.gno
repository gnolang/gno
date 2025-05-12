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
	bid = cross(boards2.CreateBoard)("test-board", false)
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "body")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.EditReply)(bid, tid, rid, "")
}

// Error:
// body is empty
