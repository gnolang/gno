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
	rid, tid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "body")
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	boards2.FlagReply(cross, bid, tid, rid, "")
}

// Error:
// unauthorized
