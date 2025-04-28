package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "body")

	testing.SetOriginCaller(user)
}

func main() {
	cross(boards2.EditReply)(bid, tid, rid, "new body")
}

// Error:
// only the reply creator is allowed to edit it
