package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	bid boards.ID
	pid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.FlagThread(cross, bid, pid, "")
}

// Error:
// flagging reason is required
