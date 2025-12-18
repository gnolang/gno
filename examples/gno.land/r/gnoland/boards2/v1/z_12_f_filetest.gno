package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	bid      boards.ID
	tid, rid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "body")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.EditReply(cross, bid, tid, rid, "new body")
}

// Error:
// only the reply creator is allowed to edit it
